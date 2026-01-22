import { notFound } from "next/navigation";
import { TagForm } from "@/components/forms/tag-form";
import { getTag } from "@/lib/api/tags";

interface TagDetailPageProps {
  params: Promise<{ id: string }>;
}

export default async function TagDetailPage({
  params,
}: TagDetailPageProps) {
  const { id } = await params;
  const tagId = parseInt(id, 10);

  if (isNaN(tagId)) {
    notFound();
  }

  try {
    const tag = await getTag(tagId);

    return (
      <div className="max-w-2xl">
        <TagForm tag={tag} />
      </div>
    );
  } catch {
    notFound();
  }
}
