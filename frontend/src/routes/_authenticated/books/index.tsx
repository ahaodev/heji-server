import { createFileRoute } from '@tanstack/react-router'
import { Books } from '@/features/billing/books'

export const Route = createFileRoute('/_authenticated/books/')({
  component: Books,
})
