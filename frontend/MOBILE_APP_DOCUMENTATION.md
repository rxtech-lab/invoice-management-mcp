# Invoice Management Mobile App Documentation

This document provides comprehensive guidance for implementing a mobile app (iOS/Swift) that mirrors the functionality of the Invoice Management web application.

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Authentication](#2-authentication)
3. [REST API Reference](#3-rest-api-reference)
4. [Data Models](#4-data-models)
5. [Screen Implementation Guide](#5-screen-implementation-guide)
6. [Form Validation Rules](#6-form-validation-rules)
7. [Currency & Exchange Rate](#7-currency--exchange-rate)
8. [File Upload](#8-file-upload)
9. [Error Handling](#9-error-handling)

---

## 1. Architecture Overview

### System Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Mobile App    │────▶│   REST API      │────▶│   Database      │
│   (iOS/Swift)   │     │   (Go/Fiber)    │     │   (Turso/SQLite)│
└─────────────────┘     └─────────────────┘     └─────────────────┘
        │                       │
        │                       ▼
        │               ┌─────────────────┐
        │               │   S3 Storage    │
        │               │   (File Upload) │
        │               └─────────────────┘
        │
        ▼
┌─────────────────┐
│   OAuth 2.0     │
│   (auth.rxlab)  │
└─────────────────┘
```

### Key Components

- **Backend API**: Go with Fiber framework at `https://api.example.com`
- **Authentication**: OAuth 2.0 with OIDC via `auth.rxlab.app`
- **File Storage**: S3-compatible storage for invoice PDFs
- **AI Agent**: MCP-based AI for automatic invoice creation from PDFs

### Base URL Configuration

```swift
struct AppConfig {
    static let apiBaseURL = "https://api.example.com"
    static let frontendBaseURL = "https://app.example.com" // For new API routes
    static let authIssuer = "https://auth.rxlab.app"
}
```

---

## 2. Authentication

### OAuth 2.0 Configuration

| Parameter | Value |
|-----------|-------|
| Issuer | `https://auth.rxlab.app` |
| Authorization Endpoint | `{issuer}/authorize` |
| Token Endpoint | `{issuer}/api/oauth/token` |
| JWKS Endpoint | `{issuer}/.well-known/jwks.json` |
| Scopes | `openid email profile offline_access` |
| Token Auth Method | `client_secret_post` |
| Supported Algorithms | RS256, RS384, RS512, ES256, ES384, ES512 |

### AuthManager Implementation

```swift
import AuthenticationServices
import Security

enum AuthError: Error {
    case noToken
    case noRefreshToken
    case authenticationFailed(String)
    case tokenRefreshFailed(String)
}

class AuthManager: ObservableObject {
    static let shared = AuthManager()

    private let issuer = AppConfig.authIssuer
    private let clientId = "your-client-id"
    private let clientSecret = "your-client-secret"
    private let redirectUri = "your-app://callback"

    @Published var isAuthenticated = false
    @Published var currentUser: User?

    // MARK: - Keychain Storage

    private let keychainService = "com.yourapp.invoice"

    var accessToken: String? {
        get { KeychainHelper.read(service: keychainService, key: "accessToken") }
        set { KeychainHelper.save(service: keychainService, key: "accessToken", value: newValue) }
    }

    var refreshToken: String? {
        get { KeychainHelper.read(service: keychainService, key: "refreshToken") }
        set { KeychainHelper.save(service: keychainService, key: "refreshToken", value: newValue) }
    }

    var tokenExpiry: Date? {
        get { KeychainHelper.readDate(service: keychainService, key: "tokenExpiry") }
        set { KeychainHelper.saveDate(service: keychainService, key: "tokenExpiry", value: newValue) }
    }

    var isTokenExpired: Bool {
        guard let expiry = tokenExpiry else { return true }
        // Refresh 5 minutes before actual expiry
        return Date() >= expiry.addingTimeInterval(-300)
    }

    // MARK: - OAuth Flow

    func login(presentationAnchor: ASPresentationAnchor) async throws {
        let state = UUID().uuidString
        let codeVerifier = generateCodeVerifier()
        let codeChallenge = generateCodeChallenge(from: codeVerifier)

        var components = URLComponents(string: "\(issuer)/authorize")!
        components.queryItems = [
            URLQueryItem(name: "client_id", value: clientId),
            URLQueryItem(name: "redirect_uri", value: redirectUri),
            URLQueryItem(name: "response_type", value: "code"),
            URLQueryItem(name: "scope", value: "openid email profile offline_access"),
            URLQueryItem(name: "state", value: state),
            URLQueryItem(name: "code_challenge", value: codeChallenge),
            URLQueryItem(name: "code_challenge_method", value: "S256"),
        ]

        let authURL = components.url!
        let callbackScheme = "your-app"

        let callbackURL = try await withCheckedThrowingContinuation { continuation in
            let session = ASWebAuthenticationSession(
                url: authURL,
                callbackURLScheme: callbackScheme
            ) { callbackURL, error in
                if let error = error {
                    continuation.resume(throwing: AuthError.authenticationFailed(error.localizedDescription))
                } else if let callbackURL = callbackURL {
                    continuation.resume(returning: callbackURL)
                } else {
                    continuation.resume(throwing: AuthError.authenticationFailed("No callback URL"))
                }
            }
            session.presentationContextProvider = ASWebAuthenticationPresentationContextProviding(anchor: presentationAnchor)
            session.prefersEphemeralWebBrowserSession = false
            session.start()
        }

        // Extract authorization code from callback
        let components2 = URLComponents(url: callbackURL, resolvingAgainstBaseURL: false)!
        guard let code = components2.queryItems?.first(where: { $0.name == "code" })?.value else {
            throw AuthError.authenticationFailed("No authorization code")
        }

        // Exchange code for tokens
        try await exchangeCodeForTokens(code: code, codeVerifier: codeVerifier)

        await MainActor.run {
            self.isAuthenticated = true
        }
    }

    private func exchangeCodeForTokens(code: String, codeVerifier: String) async throws {
        var request = URLRequest(url: URL(string: "\(issuer)/api/oauth/token")!)
        request.httpMethod = "POST"
        request.setValue("application/x-www-form-urlencoded", forHTTPHeaderField: "Content-Type")

        let body = [
            "grant_type": "authorization_code",
            "code": code,
            "redirect_uri": redirectUri,
            "client_id": clientId,
            "client_secret": clientSecret,
            "code_verifier": codeVerifier,
        ].map { "\($0.key)=\($0.value.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed)!)" }
         .joined(separator: "&")

        request.httpBody = body.data(using: .utf8)

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            throw AuthError.authenticationFailed("Token exchange failed")
        }

        let tokens = try JSONDecoder().decode(TokenResponse.self, from: data)

        self.accessToken = tokens.accessToken
        self.refreshToken = tokens.refreshToken
        self.tokenExpiry = Date().addingTimeInterval(TimeInterval(tokens.expiresIn))
    }

    // MARK: - Token Refresh

    func refreshAccessToken() async throws {
        guard let refresh = refreshToken else {
            throw AuthError.noRefreshToken
        }

        var request = URLRequest(url: URL(string: "\(issuer)/api/oauth/token")!)
        request.httpMethod = "POST"
        request.setValue("application/x-www-form-urlencoded", forHTTPHeaderField: "Content-Type")

        let body = [
            "grant_type": "refresh_token",
            "refresh_token": refresh,
            "client_id": clientId,
            "client_secret": clientSecret,
        ].map { "\($0.key)=\($0.value.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed)!)" }
         .joined(separator: "&")

        request.httpBody = body.data(using: .utf8)

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            // Refresh failed - clear tokens and require re-login
            await logout()
            throw AuthError.tokenRefreshFailed("Refresh token expired")
        }

        let tokens = try JSONDecoder().decode(TokenResponse.self, from: data)

        self.accessToken = tokens.accessToken
        self.refreshToken = tokens.refreshToken ?? refresh
        self.tokenExpiry = Date().addingTimeInterval(TimeInterval(tokens.expiresIn))
    }

    func getValidToken() async throws -> String {
        if isTokenExpired {
            try await refreshAccessToken()
        }
        guard let token = accessToken else {
            throw AuthError.noToken
        }
        return token
    }

    // MARK: - Logout

    func logout() async {
        accessToken = nil
        refreshToken = nil
        tokenExpiry = nil

        await MainActor.run {
            self.isAuthenticated = false
            self.currentUser = nil
        }
    }

    // MARK: - PKCE Helpers

    private func generateCodeVerifier() -> String {
        var bytes = [UInt8](repeating: 0, count: 32)
        _ = SecRandomCopyBytes(kSecRandomDefault, bytes.count, &bytes)
        return Data(bytes).base64EncodedString()
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
            .replacingOccurrences(of: "=", with: "")
    }

    private func generateCodeChallenge(from verifier: String) -> String {
        let data = verifier.data(using: .utf8)!
        var hash = [UInt8](repeating: 0, count: Int(CC_SHA256_DIGEST_LENGTH))
        _ = data.withUnsafeBytes { CC_SHA256($0.baseAddress, CC_LONG(data.count), &hash) }
        return Data(hash).base64EncodedString()
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
            .replacingOccurrences(of: "=", with: "")
    }
}

// MARK: - Token Response

struct TokenResponse: Codable {
    let accessToken: String
    let refreshToken: String?
    let tokenType: String
    let expiresIn: Int

    enum CodingKeys: String, CodingKey {
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
        case tokenType = "token_type"
        case expiresIn = "expires_in"
    }
}

// MARK: - Keychain Helper

struct KeychainHelper {
    static func save(service: String, key: String, value: String?) {
        guard let value = value, let data = value.data(using: .utf8) else {
            delete(service: service, key: key)
            return
        }

        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
        ]

        SecItemDelete(query as CFDictionary)

        var newItem = query
        newItem[kSecValueData as String] = data
        SecItemAdd(newItem as CFDictionary, nil)
    }

    static func read(service: String, key: String) -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
        ]

        var result: AnyObject?
        SecItemCopyMatching(query as CFDictionary, &result)

        guard let data = result as? Data else { return nil }
        return String(data: data, encoding: .utf8)
    }

    static func delete(service: String, key: String) {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
        ]
        SecItemDelete(query as CFDictionary)
    }

    static func saveDate(service: String, key: String, value: Date?) {
        save(service: service, key: key, value: value?.timeIntervalSince1970.description)
    }

    static func readDate(service: String, key: String) -> Date? {
        guard let str = read(service: service, key: key),
              let interval = TimeInterval(str) else { return nil }
        return Date(timeIntervalSince1970: interval)
    }
}
```

---

## 3. REST API Reference

### API Client

```swift
enum APIError: Error {
    case invalidResponse
    case httpError(Int, String?)
    case decodingError(Error)
    case unauthorized
}

class APIClient {
    static let shared = APIClient()

    private let baseURL = AppConfig.apiBaseURL
    private let frontendURL = AppConfig.frontendBaseURL
    private let auth = AuthManager.shared
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder

    init() {
        decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        decoder.dateDecodingStrategy = .iso8601

        encoder = JSONEncoder()
        encoder.keyEncodingStrategy = .convertToSnakeCase
        encoder.dateEncodingStrategy = .iso8601
    }

    func request<T: Decodable>(
        _ endpoint: String,
        method: String = "GET",
        body: Encodable? = nil,
        queryParams: [String: String]? = nil,
        requiresAuth: Bool = true,
        usesFrontendURL: Bool = false
    ) async throws -> T {
        var url = URL(string: (usesFrontendURL ? frontendURL : baseURL) + endpoint)!

        // Add query parameters
        if let queryParams = queryParams {
            var components = URLComponents(url: url, resolvingAgainstBaseURL: false)!
            components.queryItems = queryParams.map { URLQueryItem(name: $0.key, value: $0.value) }
            url = components.url!
        }

        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        // Add auth header if required
        if requiresAuth {
            let token = try await auth.getValidToken()
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        // Add body
        if let body = body {
            request.httpBody = try encoder.encode(body)
        }

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        // Handle 401 - try refresh and retry once
        if httpResponse.statusCode == 401 && requiresAuth {
            try await auth.refreshAccessToken()
            return try await self.request(
                endpoint,
                method: method,
                body: body,
                queryParams: queryParams,
                requiresAuth: requiresAuth,
                usesFrontendURL: usesFrontendURL
            )
        }

        // Handle 204 No Content
        if httpResponse.statusCode == 204 {
            // Return empty response for Void type
            if T.self == EmptyResponse.self {
                return EmptyResponse() as! T
            }
        }

        guard (200...299).contains(httpResponse.statusCode) else {
            let errorMessage = try? JSONDecoder().decode(APIErrorResponse.self, from: data).error
            throw APIError.httpError(httpResponse.statusCode, errorMessage)
        }

        do {
            return try decoder.decode(T.self, from: data)
        } catch {
            throw APIError.decodingError(error)
        }
    }

    // Convenience methods
    func get<T: Decodable>(_ endpoint: String, queryParams: [String: String]? = nil) async throws -> T {
        try await request(endpoint, method: "GET", queryParams: queryParams)
    }

    func post<T: Decodable, B: Encodable>(_ endpoint: String, body: B) async throws -> T {
        try await request(endpoint, method: "POST", body: body)
    }

    func put<T: Decodable, B: Encodable>(_ endpoint: String, body: B) async throws -> T {
        try await request(endpoint, method: "PUT", body: body)
    }

    func patch<T: Decodable, B: Encodable>(_ endpoint: String, body: B) async throws -> T {
        try await request(endpoint, method: "PATCH", body: body)
    }

    func delete(_ endpoint: String) async throws {
        let _: EmptyResponse = try await request(endpoint, method: "DELETE")
    }
}

struct EmptyResponse: Decodable {}
struct APIErrorResponse: Decodable { let error: String }
```

### Categories API

```swift
// MARK: - Categories

extension APIClient {
    func getCategories(keyword: String? = nil, limit: Int = 50, offset: Int = 0) async throws -> PaginatedResponse<Category> {
        var params: [String: String] = [
            "limit": String(limit),
            "offset": String(offset),
        ]
        if let keyword = keyword { params["keyword"] = keyword }

        return try await get("/api/categories", queryParams: params)
    }

    func getCategory(id: Int) async throws -> Category {
        try await get("/api/categories/\(id)")
    }

    func createCategory(_ request: CreateCategoryRequest) async throws -> Category {
        try await post("/api/categories", body: request)
    }

    func updateCategory(id: Int, _ request: UpdateCategoryRequest) async throws -> Category {
        try await put("/api/categories/\(id)", body: request)
    }

    func deleteCategory(id: Int) async throws {
        try await delete("/api/categories/\(id)")
    }
}

// Usage Example:
// let categories = try await APIClient.shared.getCategories(keyword: "utilities")
// let newCategory = try await APIClient.shared.createCategory(CreateCategoryRequest(name: "Office", color: "#FF5733"))
```

### Companies API

```swift
// MARK: - Companies

extension APIClient {
    func getCompanies(keyword: String? = nil, limit: Int = 50, offset: Int = 0) async throws -> PaginatedResponse<Company> {
        var params: [String: String] = [
            "limit": String(limit),
            "offset": String(offset),
        ]
        if let keyword = keyword { params["keyword"] = keyword }

        return try await get("/api/companies", queryParams: params)
    }

    func getCompany(id: Int) async throws -> Company {
        try await get("/api/companies/\(id)")
    }

    func createCompany(_ request: CreateCompanyRequest) async throws -> Company {
        try await post("/api/companies", body: request)
    }

    func updateCompany(id: Int, _ request: UpdateCompanyRequest) async throws -> Company {
        try await put("/api/companies/\(id)", body: request)
    }

    func deleteCompany(id: Int) async throws {
        try await delete("/api/companies/\(id)")
    }
}

// Usage Example:
// let companies = try await APIClient.shared.getCompanies()
// let company = try await APIClient.shared.createCompany(CreateCompanyRequest(name: "Acme Corp", email: "contact@acme.com"))
```

### Receivers API

```swift
// MARK: - Receivers

extension APIClient {
    func getReceivers(keyword: String? = nil, limit: Int = 50, offset: Int = 0) async throws -> PaginatedResponse<Receiver> {
        var params: [String: String] = [
            "limit": String(limit),
            "offset": String(offset),
        ]
        if let keyword = keyword { params["keyword"] = keyword }

        return try await get("/api/receivers", queryParams: params)
    }

    func getReceiver(id: Int) async throws -> Receiver {
        try await get("/api/receivers/\(id)")
    }

    func createReceiver(_ request: CreateReceiverRequest) async throws -> Receiver {
        try await post("/api/receivers", body: request)
    }

    func updateReceiver(id: Int, _ request: UpdateReceiverRequest) async throws -> Receiver {
        try await put("/api/receivers/\(id)", body: request)
    }

    func deleteReceiver(id: Int) async throws {
        try await delete("/api/receivers/\(id)")
    }
}

// Usage Example:
// let receivers = try await APIClient.shared.getReceivers()
// let receiver = try await APIClient.shared.createReceiver(CreateReceiverRequest(name: "John Doe", isOrganization: false))
```

### Invoices API

```swift
// MARK: - Invoices

extension APIClient {
    func getInvoices(
        keyword: String? = nil,
        categoryId: Int? = nil,
        companyId: Int? = nil,
        receiverId: Int? = nil,
        status: InvoiceStatus? = nil,
        sortBy: String = "created_at",
        sortOrder: String = "desc",
        limit: Int = 50,
        offset: Int = 0
    ) async throws -> PaginatedResponse<Invoice> {
        var params: [String: String] = [
            "sort_by": sortBy,
            "sort_order": sortOrder,
            "limit": String(limit),
            "offset": String(offset),
        ]
        if let keyword = keyword { params["keyword"] = keyword }
        if let categoryId = categoryId { params["category_id"] = String(categoryId) }
        if let companyId = companyId { params["company_id"] = String(companyId) }
        if let receiverId = receiverId { params["receiver_id"] = String(receiverId) }
        if let status = status { params["status"] = status.rawValue }

        return try await get("/api/invoices", queryParams: params)
    }

    func getInvoice(id: Int) async throws -> Invoice {
        try await get("/api/invoices/\(id)")
    }

    func createInvoice(_ request: CreateInvoiceRequest) async throws -> Invoice {
        try await post("/api/invoices", body: request)
    }

    func updateInvoice(id: Int, _ request: UpdateInvoiceRequest) async throws -> Invoice {
        try await put("/api/invoices/\(id)", body: request)
    }

    func updateInvoiceStatus(id: Int, status: InvoiceStatus) async throws -> Invoice {
        try await patch("/api/invoices/\(id)/status", body: ["status": status.rawValue])
    }

    func deleteInvoice(id: Int) async throws {
        try await delete("/api/invoices/\(id)")
    }
}

// Usage Example:
// let invoices = try await APIClient.shared.getInvoices(status: .unpaid, sortBy: "due_date")
// let invoice = try await APIClient.shared.getInvoice(id: 123)
```

### Invoice Items API

```swift
// MARK: - Invoice Items

extension APIClient {
    func addInvoiceItem(invoiceId: Int, _ request: CreateInvoiceItemRequest) async throws -> InvoiceItem {
        try await post("/api/invoices/\(invoiceId)/items", body: request)
    }

    func updateInvoiceItem(invoiceId: Int, itemId: Int, _ request: UpdateInvoiceItemRequest) async throws -> InvoiceItem {
        try await put("/api/invoices/\(invoiceId)/items/\(itemId)", body: request)
    }

    func deleteInvoiceItem(invoiceId: Int, itemId: Int) async throws {
        try await delete("/api/invoices/\(invoiceId)/items/\(itemId)")
    }
}

// Usage Example:
// let item = try await APIClient.shared.addInvoiceItem(invoiceId: 123, CreateInvoiceItemRequest(description: "Consulting", quantity: 10, unitPrice: 150.0))
```

### Analytics API

```swift
// MARK: - Analytics

extension APIClient {
    func getAnalyticsSummary(period: AnalyticsPeriod = .oneMonth) async throws -> AnalyticsSummary {
        try await get("/api/analytics/summary", queryParams: ["period": period.rawValue])
    }

    func getAnalyticsByCategory(period: AnalyticsPeriod = .oneMonth) async throws -> AnalyticsByGroup {
        try await get("/api/analytics/by-category", queryParams: ["period": period.rawValue])
    }

    func getAnalyticsByCompany(period: AnalyticsPeriod = .oneMonth) async throws -> AnalyticsByGroup {
        try await get("/api/analytics/by-company", queryParams: ["period": period.rawValue])
    }

    func getAnalyticsByReceiver(period: AnalyticsPeriod = .oneMonth) async throws -> AnalyticsByGroup {
        try await get("/api/analytics/by-receiver", queryParams: ["period": period.rawValue])
    }
}

// Usage Example:
// let summary = try await APIClient.shared.getAnalyticsSummary(period: .oneYear)
// let byCategory = try await APIClient.shared.getAnalyticsByCategory()
```

### Exchange Rate API (Public - No Auth Required)

```swift
// MARK: - Exchange Rate

extension APIClient {
    func getExchangeRate(from: String, to: String) async throws -> ExchangeRateResponse {
        try await request(
            "/api/exchange-rate",
            queryParams: ["from": from, "to": to],
            requiresAuth: false,
            usesFrontendURL: true
        )
    }
}

struct ExchangeRateResponse: Decodable {
    let success: Bool
    let data: ExchangeRate?
    let error: String?
}

struct ExchangeRate: Decodable {
    let from: String
    let to: String
    let rate: Double
    let date: String
}

// Usage Example:
// let rate = try await APIClient.shared.getExchangeRate(from: "USD", to: "EUR")
// let convertedAmount = amount * rate.data!.rate
```

### AI Invoice Agent API (SSE - OAuth Required)

```swift
// MARK: - AI Invoice Agent

struct AgentProgress: Decodable {
    let status: String // "idle", "calling", "complete", "error"
    let toolName: String?
    let message: String
}

class InvoiceAgentClient {
    private let auth = AuthManager.shared
    private let baseURL = AppConfig.frontendBaseURL

    func createInvoiceWithAgent(fileUrl: String) -> AsyncThrowingStream<AgentProgress, Error> {
        AsyncThrowingStream { continuation in
            Task {
                do {
                    let token = try await auth.getValidToken()

                    var request = URLRequest(url: URL(string: "\(baseURL)/api/invoices/agent")!)
                    request.httpMethod = "POST"
                    request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
                    request.setValue("application/json", forHTTPHeaderField: "Content-Type")

                    let body = ["file_url": fileUrl]
                    request.httpBody = try JSONEncoder().encode(body)

                    let (bytes, response) = try await URLSession.shared.bytes(for: request)

                    guard let httpResponse = response as? HTTPURLResponse else {
                        continuation.finish(throwing: APIError.invalidResponse)
                        return
                    }

                    if httpResponse.statusCode == 401 {
                        continuation.finish(throwing: APIError.unauthorized)
                        return
                    }

                    guard httpResponse.statusCode == 200 else {
                        continuation.finish(throwing: APIError.httpError(httpResponse.statusCode, nil))
                        return
                    }

                    // Parse SSE stream
                    var buffer = ""
                    for try await byte in bytes {
                        buffer.append(Character(UnicodeScalar(byte)))

                        // Look for complete events (double newline)
                        while let range = buffer.range(of: "\n\n") {
                            let eventStr = String(buffer[..<range.lowerBound])
                            buffer = String(buffer[range.upperBound...])

                            if let progress = parseSSEEvent(eventStr) {
                                continuation.yield(progress)

                                if progress.status == "complete" || progress.status == "error" {
                                    continuation.finish()
                                    return
                                }
                            }
                        }
                    }

                    continuation.finish()
                } catch {
                    continuation.finish(throwing: error)
                }
            }
        }
    }

    private func parseSSEEvent(_ eventStr: String) -> AgentProgress? {
        let lines = eventStr.components(separatedBy: "\n")
        var data: String?

        for line in lines {
            if line.hasPrefix("data: ") {
                data = String(line.dropFirst(6))
            }
        }

        guard let data = data,
              let jsonData = data.data(using: .utf8),
              let progress = try? JSONDecoder().decode(AgentProgress.self, from: jsonData) else {
            return nil
        }

        return progress
    }
}

// Usage Example:
// let agent = InvoiceAgentClient()
// for try await progress in agent.createInvoiceWithAgent(fileUrl: "https://storage.example.com/invoice.pdf") {
//     print("Status: \(progress.status), Message: \(progress.message)")
// }
```

---

## 4. Data Models

### Core Models

```swift
// MARK: - Invoice Status

enum InvoiceStatus: String, Codable, CaseIterable {
    case paid
    case unpaid
    case overdue

    var displayName: String {
        rawValue.capitalized
    }

    var color: Color {
        switch self {
        case .paid: return .green
        case .unpaid: return .orange
        case .overdue: return .red
        }
    }
}

// MARK: - Category

struct Category: Codable, Identifiable {
    let id: Int
    let userId: String
    let name: String
    let description: String?
    let color: String?
    let createdAt: Date
    let updatedAt: Date
}

struct CreateCategoryRequest: Codable {
    let name: String
    let description: String?
    let color: String?

    init(name: String, description: String? = nil, color: String? = nil) {
        self.name = name
        self.description = description
        self.color = color
    }
}

struct UpdateCategoryRequest: Codable {
    let name: String?
    let description: String?
    let color: String?
}

// MARK: - Company

struct Company: Codable, Identifiable {
    let id: Int
    let userId: String
    let name: String
    let address: String?
    let email: String?
    let phone: String?
    let website: String?
    let taxId: String?
    let notes: String?
    let createdAt: Date
    let updatedAt: Date
}

struct CreateCompanyRequest: Codable {
    let name: String
    let address: String?
    let email: String?
    let phone: String?
    let website: String?
    let taxId: String?
    let notes: String?

    init(name: String, address: String? = nil, email: String? = nil, phone: String? = nil, website: String? = nil, taxId: String? = nil, notes: String? = nil) {
        self.name = name
        self.address = address
        self.email = email
        self.phone = phone
        self.website = website
        self.taxId = taxId
        self.notes = notes
    }
}

struct UpdateCompanyRequest: Codable {
    let name: String?
    let address: String?
    let email: String?
    let phone: String?
    let website: String?
    let taxId: String?
    let notes: String?
}

// MARK: - Receiver

struct Receiver: Codable, Identifiable {
    let id: Int
    let userId: String
    let name: String
    let isOrganization: Bool
    let createdAt: Date
    let updatedAt: Date
}

struct CreateReceiverRequest: Codable {
    let name: String
    let isOrganization: Bool?

    init(name: String, isOrganization: Bool? = false) {
        self.name = name
        self.isOrganization = isOrganization
    }
}

struct UpdateReceiverRequest: Codable {
    let name: String?
    let isOrganization: Bool?
}

// MARK: - Invoice Item

struct InvoiceItem: Codable, Identifiable {
    let id: Int
    let invoiceId: Int
    let description: String
    let quantity: Double
    let unitPrice: Double
    let amount: Double // Calculated: quantity * unitPrice
    let createdAt: Date
    let updatedAt: Date
}

struct CreateInvoiceItemRequest: Codable {
    let description: String
    let quantity: Double?
    let unitPrice: Double?

    init(description: String, quantity: Double? = 1.0, unitPrice: Double? = 0.0) {
        self.description = description
        self.quantity = quantity
        self.unitPrice = unitPrice
    }
}

struct UpdateInvoiceItemRequest: Codable {
    let description: String?
    let quantity: Double?
    let unitPrice: Double?
}

// MARK: - Invoice

struct Invoice: Codable, Identifiable {
    let id: Int
    let userId: String
    let title: String
    let description: String?
    let invoiceStartedAt: Date?
    let invoiceEndedAt: Date?
    let amount: Double // Calculated from items
    let currency: String
    let categoryId: Int?
    let category: Category?
    let companyId: Int?
    let company: Company?
    let receiverId: Int?
    let receiver: Receiver?
    let items: [InvoiceItem]
    let originalDownloadLink: String?
    let tags: [String]?
    let status: InvoiceStatus
    let dueDate: Date?
    let createdAt: Date
    let updatedAt: Date
}

struct CreateInvoiceRequest: Codable {
    let title: String
    let description: String?
    let invoiceStartedAt: Date?
    let invoiceEndedAt: Date?
    let currency: String?
    let categoryId: Int?
    let companyId: Int?
    let receiverId: Int?
    let originalDownloadLink: String?
    let tags: [String]?
    let status: InvoiceStatus?
    let dueDate: Date?
    let items: [CreateInvoiceItemRequest]?

    init(
        title: String,
        description: String? = nil,
        invoiceStartedAt: Date? = nil,
        invoiceEndedAt: Date? = nil,
        currency: String? = "USD",
        categoryId: Int? = nil,
        companyId: Int? = nil,
        receiverId: Int? = nil,
        originalDownloadLink: String? = nil,
        tags: [String]? = nil,
        status: InvoiceStatus? = .unpaid,
        dueDate: Date? = nil,
        items: [CreateInvoiceItemRequest]? = nil
    ) {
        self.title = title
        self.description = description
        self.invoiceStartedAt = invoiceStartedAt
        self.invoiceEndedAt = invoiceEndedAt
        self.currency = currency
        self.categoryId = categoryId
        self.companyId = companyId
        self.receiverId = receiverId
        self.originalDownloadLink = originalDownloadLink
        self.tags = tags
        self.status = status
        self.dueDate = dueDate
        self.items = items
    }
}

struct UpdateInvoiceRequest: Codable {
    let title: String?
    let description: String?
    let invoiceStartedAt: Date?
    let invoiceEndedAt: Date?
    let currency: String?
    let categoryId: Int?
    let companyId: Int?
    let receiverId: Int?
    let originalDownloadLink: String?
    let tags: [String]?
    let status: InvoiceStatus?
    let dueDate: Date?
}

// MARK: - Analytics

enum AnalyticsPeriod: String, CaseIterable {
    case sevenDays = "7d"
    case oneMonth = "1m"
    case oneYear = "1y"

    var displayName: String {
        switch self {
        case .sevenDays: return "7 Days"
        case .oneMonth: return "1 Month"
        case .oneYear: return "1 Year"
        }
    }
}

struct AnalyticsSummary: Codable {
    let period: String
    let startDate: Date
    let endDate: Date
    let totalAmount: Double
    let paidAmount: Double
    let unpaidAmount: Double
    let overdueAmount: Double
    let invoiceCount: Int
    let paidCount: Int
    let unpaidCount: Int
    let overdueCount: Int
}

struct AnalyticsGroupItem: Codable, Identifiable {
    let id: Int
    let name: String
    let color: String?
    let totalAmount: Double
    let paidAmount: Double
    let unpaidAmount: Double
    let invoiceCount: Int
}

struct AnalyticsByGroup: Codable {
    let period: String
    let startDate: Date
    let endDate: Date
    let items: [AnalyticsGroupItem]
    let uncategorized: AnalyticsGroupItem?
}

// MARK: - Paginated Response

struct PaginatedResponse<T: Codable>: Codable {
    let data: [T]
    let total: Int
    let limit: Int
    let offset: Int
}
```

---

## 5. Screen Implementation Guide

### Dashboard Screen

**Features:**
- Summary cards showing total, paid, unpaid, and overdue amounts
- Period selector (7d, 1m, 1y)
- Spending trend chart
- Category breakdown pie chart
- Company breakdown pie chart
- Receiver breakdown pie chart

```swift
struct DashboardView: View {
    @StateObject private var viewModel = DashboardViewModel()

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 20) {
                    // Period Selector
                    PeriodSelectorView(selectedPeriod: $viewModel.selectedPeriod)

                    // Summary Cards
                    LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 16) {
                        SummaryCard(title: "Total", amount: viewModel.summary?.totalAmount ?? 0, count: viewModel.summary?.invoiceCount ?? 0)
                        SummaryCard(title: "Paid", amount: viewModel.summary?.paidAmount ?? 0, count: viewModel.summary?.paidCount ?? 0, color: .green)
                        SummaryCard(title: "Unpaid", amount: viewModel.summary?.unpaidAmount ?? 0, count: viewModel.summary?.unpaidCount ?? 0, color: .orange)
                        SummaryCard(title: "Overdue", amount: viewModel.summary?.overdueAmount ?? 0, count: viewModel.summary?.overdueCount ?? 0, color: .red)
                    }

                    // Charts
                    if let byCategory = viewModel.byCategory {
                        PieChartView(title: "By Category", data: byCategory.items)
                    }

                    if let byCompany = viewModel.byCompany {
                        PieChartView(title: "By Company", data: byCompany.items)
                    }
                }
                .padding()
            }
            .navigationTitle("Dashboard")
            .refreshable {
                await viewModel.loadData()
            }
        }
        .task {
            await viewModel.loadData()
        }
    }
}

@MainActor
class DashboardViewModel: ObservableObject {
    @Published var selectedPeriod: AnalyticsPeriod = .oneMonth
    @Published var summary: AnalyticsSummary?
    @Published var byCategory: AnalyticsByGroup?
    @Published var byCompany: AnalyticsByGroup?
    @Published var byReceiver: AnalyticsByGroup?
    @Published var isLoading = false

    func loadData() async {
        isLoading = true
        defer { isLoading = false }

        async let summaryTask = APIClient.shared.getAnalyticsSummary(period: selectedPeriod)
        async let categoryTask = APIClient.shared.getAnalyticsByCategory(period: selectedPeriod)
        async let companyTask = APIClient.shared.getAnalyticsByCompany(period: selectedPeriod)
        async let receiverTask = APIClient.shared.getAnalyticsByReceiver(period: selectedPeriod)

        do {
            let (s, c, co, r) = try await (summaryTask, categoryTask, companyTask, receiverTask)
            self.summary = s
            self.byCategory = c
            self.byCompany = co
            self.byReceiver = r
        } catch {
            print("Error loading dashboard: \(error)")
        }
    }
}
```

### Invoices Screen

**Features:**
- List view with pull-to-refresh
- Search and filter UI (keyword, category, company, receiver, status)
- Status badges
- Swipe actions (edit, delete, change status)
- AI invoice creation button

```swift
struct InvoicesView: View {
    @StateObject private var viewModel = InvoicesViewModel()
    @State private var showFilters = false
    @State private var showNewInvoice = false
    @State private var showAIAgent = false

    var body: some View {
        NavigationStack {
            List {
                ForEach(viewModel.invoices) { invoice in
                    NavigationLink(destination: InvoiceDetailView(invoice: invoice)) {
                        InvoiceRowView(invoice: invoice)
                    }
                    .swipeActions(edge: .trailing) {
                        Button(role: .destructive) {
                            Task { await viewModel.deleteInvoice(invoice) }
                        } label: {
                            Label("Delete", systemImage: "trash")
                        }
                    }
                    .swipeActions(edge: .leading) {
                        Button {
                            Task { await viewModel.toggleStatus(invoice) }
                        } label: {
                            Label(invoice.status == .paid ? "Mark Unpaid" : "Mark Paid",
                                  systemImage: invoice.status == .paid ? "xmark.circle" : "checkmark.circle")
                        }
                        .tint(invoice.status == .paid ? .orange : .green)
                    }
                }

                // Load more
                if viewModel.hasMore {
                    ProgressView()
                        .task { await viewModel.loadMore() }
                }
            }
            .searchable(text: $viewModel.searchText)
            .navigationTitle("Invoices")
            .toolbar {
                ToolbarItem(placement: .navigationBarLeading) {
                    Button {
                        showFilters.toggle()
                    } label: {
                        Image(systemName: "line.3.horizontal.decrease.circle")
                    }
                }

                ToolbarItem(placement: .navigationBarTrailing) {
                    Menu {
                        Button {
                            showNewInvoice = true
                        } label: {
                            Label("New Invoice", systemImage: "doc.badge.plus")
                        }

                        Button {
                            showAIAgent = true
                        } label: {
                            Label("Create with AI", systemImage: "sparkles")
                        }
                    } label: {
                        Image(systemName: "plus")
                    }
                }
            }
            .refreshable {
                await viewModel.refresh()
            }
            .sheet(isPresented: $showFilters) {
                InvoiceFiltersSheet(viewModel: viewModel)
            }
            .sheet(isPresented: $showNewInvoice) {
                InvoiceFormView(mode: .create)
            }
            .sheet(isPresented: $showAIAgent) {
                AIAgentView()
            }
        }
        .task {
            await viewModel.loadInvoices()
        }
    }
}

@MainActor
class InvoicesViewModel: ObservableObject {
    @Published var invoices: [Invoice] = []
    @Published var searchText = ""
    @Published var selectedCategory: Category?
    @Published var selectedCompany: Company?
    @Published var selectedReceiver: Receiver?
    @Published var selectedStatus: InvoiceStatus?
    @Published var isLoading = false
    @Published var hasMore = true

    private var offset = 0
    private let limit = 20

    func loadInvoices() async {
        isLoading = true
        offset = 0

        do {
            let response = try await APIClient.shared.getInvoices(
                keyword: searchText.isEmpty ? nil : searchText,
                categoryId: selectedCategory?.id,
                companyId: selectedCompany?.id,
                receiverId: selectedReceiver?.id,
                status: selectedStatus,
                limit: limit,
                offset: offset
            )
            invoices = response.data
            hasMore = response.data.count == limit
        } catch {
            print("Error loading invoices: \(error)")
        }

        isLoading = false
    }

    func loadMore() async {
        offset += limit

        do {
            let response = try await APIClient.shared.getInvoices(
                keyword: searchText.isEmpty ? nil : searchText,
                categoryId: selectedCategory?.id,
                companyId: selectedCompany?.id,
                receiverId: selectedReceiver?.id,
                status: selectedStatus,
                limit: limit,
                offset: offset
            )
            invoices.append(contentsOf: response.data)
            hasMore = response.data.count == limit
        } catch {
            print("Error loading more: \(error)")
        }
    }

    func refresh() async {
        await loadInvoices()
    }

    func deleteInvoice(_ invoice: Invoice) async {
        do {
            try await APIClient.shared.deleteInvoice(id: invoice.id)
            invoices.removeAll { $0.id == invoice.id }
        } catch {
            print("Error deleting: \(error)")
        }
    }

    func toggleStatus(_ invoice: Invoice) async {
        let newStatus: InvoiceStatus = invoice.status == .paid ? .unpaid : .paid
        do {
            let updated = try await APIClient.shared.updateInvoiceStatus(id: invoice.id, status: newStatus)
            if let index = invoices.firstIndex(where: { $0.id == invoice.id }) {
                invoices[index] = updated
            }
        } catch {
            print("Error updating status: \(error)")
        }
    }
}
```

### Invoice Detail/Edit Screen

**Features:**
- Form fields (title, description, dates, currency, status)
- Category/Company/Receiver pickers
- Invoice items table with inline editing
- Currency conversion display
- Download original file link

```swift
struct InvoiceDetailView: View {
    @StateObject private var viewModel: InvoiceDetailViewModel
    @Environment(\.dismiss) private var dismiss

    init(invoice: Invoice) {
        _viewModel = StateObject(wrappedValue: InvoiceDetailViewModel(invoice: invoice))
    }

    var body: some View {
        Form {
            Section("Details") {
                TextField("Title", text: $viewModel.title)
                TextField("Description", text: $viewModel.description, axis: .vertical)

                Picker("Status", selection: $viewModel.status) {
                    ForEach(InvoiceStatus.allCases, id: \.self) { status in
                        Text(status.displayName).tag(status)
                    }
                }

                Picker("Currency", selection: $viewModel.currency) {
                    ForEach(SupportedCurrencies.all, id: \.code) { currency in
                        Text("\(currency.code) - \(currency.name)").tag(currency.code)
                    }
                }
            }

            Section("Dates") {
                DatePicker("Due Date", selection: $viewModel.dueDate, displayedComponents: .date)
                DatePicker("Start Date", selection: $viewModel.startDate, displayedComponents: .date)
                DatePicker("End Date", selection: $viewModel.endDate, displayedComponents: .date)
            }

            Section("Links") {
                NavigationLink(destination: CategoryPickerView(selected: $viewModel.selectedCategory)) {
                    HStack {
                        Text("Category")
                        Spacer()
                        Text(viewModel.selectedCategory?.name ?? "None")
                            .foregroundColor(.secondary)
                    }
                }

                NavigationLink(destination: CompanyPickerView(selected: $viewModel.selectedCompany)) {
                    HStack {
                        Text("Company")
                        Spacer()
                        Text(viewModel.selectedCompany?.name ?? "None")
                            .foregroundColor(.secondary)
                    }
                }

                NavigationLink(destination: ReceiverPickerView(selected: $viewModel.selectedReceiver)) {
                    HStack {
                        Text("Receiver")
                        Spacer()
                        Text(viewModel.selectedReceiver?.name ?? "None")
                            .foregroundColor(.secondary)
                    }
                }
            }

            Section("Items") {
                ForEach(viewModel.items) { item in
                    InvoiceItemRow(item: item, onUpdate: { updated in
                        Task { await viewModel.updateItem(updated) }
                    }, onDelete: {
                        Task { await viewModel.deleteItem(item) }
                    })
                }

                Button {
                    viewModel.addNewItem()
                } label: {
                    Label("Add Item", systemImage: "plus.circle")
                }
            }

            Section {
                HStack {
                    Text("Total")
                        .font(.headline)
                    Spacer()
                    Text(viewModel.formattedTotal)
                        .font(.headline)
                }
            }

            if let downloadLink = viewModel.originalDownloadLink, !downloadLink.isEmpty {
                Section("Attachments") {
                    Link(destination: URL(string: downloadLink)!) {
                        Label("Download Original", systemImage: "arrow.down.doc")
                    }
                }
            }
        }
        .navigationTitle("Invoice")
        .toolbar {
            ToolbarItem(placement: .navigationBarTrailing) {
                Button("Save") {
                    Task {
                        await viewModel.save()
                        dismiss()
                    }
                }
                .disabled(!viewModel.hasChanges)
            }
        }
    }
}
```

### Categories Screen

```swift
struct CategoriesView: View {
    @StateObject private var viewModel = CategoriesViewModel()
    @State private var showNewCategory = false

    var body: some View {
        NavigationStack {
            List {
                ForEach(viewModel.categories) { category in
                    NavigationLink(destination: CategoryFormView(category: category)) {
                        HStack {
                            Circle()
                                .fill(Color(hex: category.color ?? "#888888"))
                                .frame(width: 20, height: 20)
                            VStack(alignment: .leading) {
                                Text(category.name)
                                    .font(.headline)
                                if let description = category.description {
                                    Text(description)
                                        .font(.caption)
                                        .foregroundColor(.secondary)
                                }
                            }
                        }
                    }
                    .swipeActions {
                        Button(role: .destructive) {
                            Task { await viewModel.delete(category) }
                        } label: {
                            Label("Delete", systemImage: "trash")
                        }
                    }
                }
            }
            .searchable(text: $viewModel.searchText)
            .navigationTitle("Categories")
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button {
                        showNewCategory = true
                    } label: {
                        Image(systemName: "plus")
                    }
                }
            }
            .refreshable {
                await viewModel.load()
            }
            .sheet(isPresented: $showNewCategory) {
                CategoryFormView(category: nil)
            }
        }
        .task {
            await viewModel.load()
        }
    }
}
```

### Companies Screen

```swift
struct CompaniesView: View {
    @StateObject private var viewModel = CompaniesViewModel()
    @State private var showNewCompany = false

    var body: some View {
        NavigationStack {
            List {
                ForEach(viewModel.companies) { company in
                    NavigationLink(destination: CompanyFormView(company: company)) {
                        VStack(alignment: .leading, spacing: 4) {
                            Text(company.name)
                                .font(.headline)
                            if let email = company.email, !email.isEmpty {
                                Text(email)
                                    .font(.caption)
                                    .foregroundColor(.secondary)
                            }
                            if let phone = company.phone, !phone.isEmpty {
                                Text(phone)
                                    .font(.caption)
                                    .foregroundColor(.secondary)
                            }
                        }
                    }
                    .swipeActions {
                        Button(role: .destructive) {
                            Task { await viewModel.delete(company) }
                        } label: {
                            Label("Delete", systemImage: "trash")
                        }
                    }
                }
            }
            .searchable(text: $viewModel.searchText)
            .navigationTitle("Companies")
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button {
                        showNewCompany = true
                    } label: {
                        Image(systemName: "plus")
                    }
                }
            }
            .refreshable {
                await viewModel.load()
            }
            .sheet(isPresented: $showNewCompany) {
                CompanyFormView(company: nil)
            }
        }
        .task {
            await viewModel.load()
        }
    }
}
```

### Receivers Screen

```swift
struct ReceiversView: View {
    @StateObject private var viewModel = ReceiversViewModel()
    @State private var showNewReceiver = false

    var body: some View {
        NavigationStack {
            List {
                ForEach(viewModel.receivers) { receiver in
                    NavigationLink(destination: ReceiverFormView(receiver: receiver)) {
                        HStack {
                            Image(systemName: receiver.isOrganization ? "building.2" : "person")
                                .foregroundColor(.secondary)
                            VStack(alignment: .leading) {
                                Text(receiver.name)
                                    .font(.headline)
                                Text(receiver.isOrganization ? "Organization" : "Individual")
                                    .font(.caption)
                                    .foregroundColor(.secondary)
                            }
                        }
                    }
                    .swipeActions {
                        Button(role: .destructive) {
                            Task { await viewModel.delete(receiver) }
                        } label: {
                            Label("Delete", systemImage: "trash")
                        }
                    }
                }
            }
            .searchable(text: $viewModel.searchText)
            .navigationTitle("Receivers")
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button {
                        showNewReceiver = true
                    } label: {
                        Image(systemName: "plus")
                    }
                }
            }
            .refreshable {
                await viewModel.load()
            }
            .sheet(isPresented: $showNewReceiver) {
                ReceiverFormView(receiver: nil)
            }
        }
        .task {
            await viewModel.load()
        }
    }
}
```

---

## 6. Form Validation Rules

### Invoice Validation

```swift
struct InvoiceValidator {
    static func validate(_ invoice: CreateInvoiceRequest) -> [ValidationError] {
        var errors: [ValidationError] = []

        // Title is required
        if invoice.title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            errors.append(ValidationError(field: "title", message: "Title is required"))
        }

        // Currency must be valid
        if let currency = invoice.currency {
            if !SupportedCurrencies.all.contains(where: { $0.code == currency }) {
                errors.append(ValidationError(field: "currency", message: "Invalid currency code"))
            }
        }

        // Due date should not be in the past (for new invoices)
        if let dueDate = invoice.dueDate, dueDate < Date() {
            errors.append(ValidationError(field: "dueDate", message: "Due date cannot be in the past"))
        }

        // Start date should be before end date
        if let start = invoice.invoiceStartedAt, let end = invoice.invoiceEndedAt, start > end {
            errors.append(ValidationError(field: "invoiceEndedAt", message: "End date must be after start date"))
        }

        return errors
    }
}
```

### Category Validation

```swift
struct CategoryValidator {
    static func validate(_ category: CreateCategoryRequest) -> [ValidationError] {
        var errors: [ValidationError] = []

        // Name is required
        if category.name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            errors.append(ValidationError(field: "name", message: "Name is required"))
        }

        // Color must be valid hex if provided
        if let color = category.color, !color.isEmpty {
            let hexRegex = "^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$"
            if color.range(of: hexRegex, options: .regularExpression) == nil {
                errors.append(ValidationError(field: "color", message: "Invalid hex color format (use #RRGGBB)"))
            }
        }

        return errors
    }
}
```

### Company Validation

```swift
struct CompanyValidator {
    static func validate(_ company: CreateCompanyRequest) -> [ValidationError] {
        var errors: [ValidationError] = []

        // Name is required
        if company.name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            errors.append(ValidationError(field: "name", message: "Name is required"))
        }

        // Email must be valid if provided
        if let email = company.email, !email.isEmpty {
            let emailRegex = "[A-Z0-9a-z._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,64}"
            if email.range(of: emailRegex, options: .regularExpression) == nil {
                errors.append(ValidationError(field: "email", message: "Invalid email format"))
            }
        }

        // Website must be valid URL if provided
        if let website = company.website, !website.isEmpty {
            if URL(string: website) == nil {
                errors.append(ValidationError(field: "website", message: "Invalid website URL"))
            }
        }

        return errors
    }
}
```

### Receiver Validation

```swift
struct ReceiverValidator {
    static func validate(_ receiver: CreateReceiverRequest) -> [ValidationError] {
        var errors: [ValidationError] = []

        // Name is required
        if receiver.name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            errors.append(ValidationError(field: "name", message: "Name is required"))
        }

        return errors
    }
}

struct ValidationError: Identifiable {
    let id = UUID()
    let field: String
    let message: String
}
```

---

## 7. Currency & Exchange Rate

### Supported Currencies

```swift
struct SupportedCurrencies {
    static let all: [CurrencyInfo] = [
        CurrencyInfo(code: "USD", name: "US Dollar", symbol: "$"),
        CurrencyInfo(code: "EUR", name: "Euro", symbol: "€"),
        CurrencyInfo(code: "GBP", name: "British Pound", symbol: "£"),
        CurrencyInfo(code: "JPY", name: "Japanese Yen", symbol: "¥"),
        CurrencyInfo(code: "CNY", name: "Chinese Yuan", symbol: "¥"),
        CurrencyInfo(code: "CAD", name: "Canadian Dollar", symbol: "C$"),
        CurrencyInfo(code: "AUD", name: "Australian Dollar", symbol: "A$"),
        CurrencyInfo(code: "CHF", name: "Swiss Franc", symbol: "CHF"),
        CurrencyInfo(code: "INR", name: "Indian Rupee", symbol: "₹"),
        CurrencyInfo(code: "MXN", name: "Mexican Peso", symbol: "$"),
    ]

    static func symbol(for code: String) -> String {
        all.first { $0.code == code }?.symbol ?? code
    }
}

struct CurrencyInfo {
    let code: String
    let name: String
    let symbol: String
}
```

### Currency Conversion

```swift
class CurrencyConverter: ObservableObject {
    @Published var rate: Double?
    @Published var isLoading = false
    @Published var error: String?

    private var cachedRates: [String: (rate: Double, timestamp: Date)] = [:]
    private let cacheTimeout: TimeInterval = 3600 // 1 hour

    func convert(amount: Double, from: String, to: String) async -> Double? {
        if from == to { return amount }

        // Check cache
        let cacheKey = "\(from)-\(to)"
        if let cached = cachedRates[cacheKey],
           Date().timeIntervalSince(cached.timestamp) < cacheTimeout {
            return amount * cached.rate
        }

        // Fetch new rate
        do {
            let response = try await APIClient.shared.getExchangeRate(from: from, to: to)
            if let data = response.data {
                cachedRates[cacheKey] = (data.rate, Date())
                await MainActor.run { self.rate = data.rate }
                return amount * data.rate
            }
        } catch {
            await MainActor.run { self.error = error.localizedDescription }
        }

        return nil
    }

    func formatCurrency(_ amount: Double, currency: String) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencyCode = currency
        formatter.currencySymbol = SupportedCurrencies.symbol(for: currency)
        return formatter.string(from: NSNumber(value: amount)) ?? "\(currency) \(amount)"
    }
}
```

---

## 8. File Upload

### Upload Service

```swift
class UploadService {
    static let shared = UploadService()

    private let api = APIClient.shared

    // Direct upload
    func uploadFile(data: Data, filename: String, contentType: String) async throws -> UploadResponse {
        let token = try await AuthManager.shared.getValidToken()

        let boundary = UUID().uuidString
        var request = URLRequest(url: URL(string: "\(AppConfig.apiBaseURL)/api/upload")!)
        request.httpMethod = "POST"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")

        var body = Data()
        body.append("--\(boundary)\r\n".data(using: .utf8)!)
        body.append("Content-Disposition: form-data; name=\"file\"; filename=\"\(filename)\"\r\n".data(using: .utf8)!)
        body.append("Content-Type: \(contentType)\r\n\r\n".data(using: .utf8)!)
        body.append(data)
        body.append("\r\n--\(boundary)--\r\n".data(using: .utf8)!)

        request.httpBody = body

        let (responseData, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse,
              (200...299).contains(httpResponse.statusCode) else {
            throw APIError.invalidResponse
        }

        return try JSONDecoder().decode(UploadResponse.self, from: responseData)
    }

    // Presigned URL upload (for large files)
    func getPresignedURL(filename: String, contentType: String = "application/octet-stream") async throws -> PresignedURLResponse {
        try await api.get("/api/upload/presigned", queryParams: [
            "filename": filename,
            "content_type": contentType
        ])
    }

    func uploadToPresignedURL(_ url: String, data: Data, contentType: String) async throws {
        var request = URLRequest(url: URL(string: url)!)
        request.httpMethod = "PUT"
        request.setValue(contentType, forHTTPHeaderField: "Content-Type")
        request.httpBody = data

        let (_, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse,
              (200...299).contains(httpResponse.statusCode) else {
            throw APIError.invalidResponse
        }
    }
}

struct UploadResponse: Codable {
    let key: String
    let downloadUrl: String
    let filename: String
    let size: Int
    let contentType: String
}

struct PresignedURLResponse: Codable {
    let uploadUrl: String
    let key: String
    let contentType: String
}
```

### Document Picker Integration

```swift
import UniformTypeIdentifiers

struct DocumentPicker: UIViewControllerRepresentable {
    let onPick: (URL) -> Void

    func makeUIViewController(context: Context) -> UIDocumentPickerViewController {
        let picker = UIDocumentPickerViewController(forOpeningContentTypes: [.pdf, .image])
        picker.allowsMultipleSelection = false
        picker.delegate = context.coordinator
        return picker
    }

    func updateUIViewController(_ uiViewController: UIDocumentPickerViewController, context: Context) {}

    func makeCoordinator() -> Coordinator {
        Coordinator(onPick: onPick)
    }

    class Coordinator: NSObject, UIDocumentPickerDelegate {
        let onPick: (URL) -> Void

        init(onPick: @escaping (URL) -> Void) {
            self.onPick = onPick
        }

        func documentPicker(_ controller: UIDocumentPickerViewController, didPickDocumentsAt urls: [URL]) {
            guard let url = urls.first else { return }
            onPick(url)
        }
    }
}
```

---

## 9. Error Handling

### Global Error Handler

```swift
class ErrorHandler: ObservableObject {
    static let shared = ErrorHandler()

    @Published var currentError: AppError?
    @Published var showError = false

    func handle(_ error: Error) {
        let appError: AppError

        switch error {
        case let apiError as APIError:
            switch apiError {
            case .unauthorized:
                appError = AppError(title: "Session Expired", message: "Please log in again.", action: .logout)
            case .httpError(let code, let message):
                appError = AppError(title: "Error \(code)", message: message ?? "An error occurred.", action: .dismiss)
            case .decodingError:
                appError = AppError(title: "Data Error", message: "Failed to process server response.", action: .dismiss)
            case .invalidResponse:
                appError = AppError(title: "Network Error", message: "Invalid response from server.", action: .retry)
            }
        case let authError as AuthError:
            switch authError {
            case .noToken, .noRefreshToken, .tokenRefreshFailed:
                appError = AppError(title: "Authentication Required", message: "Please log in to continue.", action: .logout)
            case .authenticationFailed(let message):
                appError = AppError(title: "Login Failed", message: message, action: .dismiss)
            }
        default:
            appError = AppError(title: "Error", message: error.localizedDescription, action: .dismiss)
        }

        DispatchQueue.main.async {
            self.currentError = appError
            self.showError = true
        }
    }
}

struct AppError: Identifiable {
    let id = UUID()
    let title: String
    let message: String
    let action: ErrorAction

    enum ErrorAction {
        case dismiss
        case retry
        case logout
    }
}

// Usage in views
struct ContentView: View {
    @ObservedObject var errorHandler = ErrorHandler.shared

    var body: some View {
        // Main content
        TabView { /* ... */ }
            .alert(errorHandler.currentError?.title ?? "Error",
                   isPresented: $errorHandler.showError) {
                switch errorHandler.currentError?.action {
                case .logout:
                    Button("Log Out") {
                        Task { await AuthManager.shared.logout() }
                    }
                case .retry:
                    Button("Retry") { /* Implement retry */ }
                    Button("Cancel", role: .cancel) {}
                default:
                    Button("OK", role: .cancel) {}
                }
            } message: {
                Text(errorHandler.currentError?.message ?? "")
            }
    }
}
```

---

## Summary

This documentation covers all the essential components for building an iOS mobile app that mirrors the Invoice Management web application:

1. **Authentication**: Complete OAuth 2.0 implementation with PKCE and token refresh
2. **API Client**: Type-safe API client with automatic authentication
3. **Data Models**: All Swift structs matching the backend models
4. **Screen Implementations**: Dashboard, Invoices, Categories, Companies, and Receivers
5. **Validation**: Form validation rules for all entities
6. **Currency**: Exchange rate integration and formatting
7. **File Upload**: Direct and presigned URL upload support
8. **Error Handling**: Global error handling with user feedback

For questions or updates, refer to the web application codebase at `frontend/` for the latest patterns and implementations.
