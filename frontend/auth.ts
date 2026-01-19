import NextAuth from "next-auth";

declare module "next-auth" {
  interface Session {
    accessToken?: string;
    error?: string;
    user: {
      id: string;
      name?: string | null;
      email?: string | null;
      image?: string | null;
    };
  }
}

export const { handlers, auth, signIn, signOut } = NextAuth({
  debug: process.env.NODE_ENV === "development",
  providers: [
    {
      id: "rxlab",
      name: "RxLab",
      type: "oidc",
      issuer: process.env.AUTH_ISSUER,
      clientId: process.env.AUTH_CLIENT_ID!,
      clientSecret: process.env.AUTH_CLIENT_SECRET!,
      client: {
        token_endpoint_auth_method: "client_secret_post",
      },
      authorization: {
        params: {
          scope: "openid email profile offline_access",
        },
      },
    },
  ],
  callbacks: {
    async jwt({ token, account }) {
      // Initial login - store all tokens
      if (account) {
        return {
          ...token,
          accessToken: account.access_token,
          refreshToken: account.refresh_token,
          expiresAt: account.expires_at,
        };
      }

      // Token still valid - return as is
      if (token.expiresAt && Date.now() < (token.expiresAt as number) * 1000) {
        return token;
      }

      // Token expired - attempt refresh
      if (!token.refreshToken) {
        console.error("No refresh token available", token);
        return { ...token, error: "RefreshTokenError" };
      }

      try {
        const response = await fetch(
          `${process.env.AUTH_ISSUER}/api/oauth/token`,
          {
            method: "POST",
            headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: new URLSearchParams({
              grant_type: "refresh_token",
              refresh_token: token.refreshToken as string,
              client_id: process.env.AUTH_CLIENT_ID!,
              client_secret: process.env.AUTH_CLIENT_SECRET!,
            }),
          }
        );

        const tokens = await response.json();
        if (!response.ok) throw tokens;

        return {
          ...token,
          accessToken: tokens.access_token,
          refreshToken: tokens.refresh_token ?? token.refreshToken,
          expiresAt: Math.floor(Date.now() / 1000) + tokens.expires_in,
          error: undefined,
        };
      } catch (error) {
        console.error("Token refresh failed:", error);
        return { ...token, error: "RefreshTokenError" };
      }
    },
    async session({ session, token }) {
      // Make access token available to the client
      session.accessToken = token.accessToken as string | undefined;
      session.error = token.error as string | undefined;
      return session;
    },
  },
  pages: {
    signIn: "/login",
  },
});
