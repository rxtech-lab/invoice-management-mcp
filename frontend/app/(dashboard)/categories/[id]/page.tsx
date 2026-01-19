import { notFound } from "next/navigation";
import { CategoryForm } from "@/components/forms/category-form";
import { getCategory } from "@/lib/api/categories";

interface CategoryDetailPageProps {
  params: Promise<{ id: string }>;
}

export default async function CategoryDetailPage({
  params,
}: CategoryDetailPageProps) {
  const { id } = await params;
  const categoryId = parseInt(id, 10);

  if (isNaN(categoryId)) {
    notFound();
  }

  try {
    const category = await getCategory(categoryId);

    return (
      <div className="max-w-2xl">
        <CategoryForm category={category} />
      </div>
    );
  } catch {
    notFound();
  }
}
