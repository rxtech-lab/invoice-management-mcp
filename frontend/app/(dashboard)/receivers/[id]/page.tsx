import { notFound } from "next/navigation";
import { ReceiverForm } from "@/components/forms/receiver-form";
import { getReceiver } from "@/lib/api/receivers";

interface ReceiverDetailPageProps {
  params: Promise<{ id: string }>;
}

export default async function ReceiverDetailPage({
  params,
}: ReceiverDetailPageProps) {
  const { id } = await params;
  const receiverId = parseInt(id, 10);

  if (isNaN(receiverId)) {
    notFound();
  }

  try {
    const receiver = await getReceiver(receiverId);

    return (
      <div className="max-w-2xl">
        <ReceiverForm receiver={receiver} />
      </div>
    );
  } catch {
    notFound();
  }
}
