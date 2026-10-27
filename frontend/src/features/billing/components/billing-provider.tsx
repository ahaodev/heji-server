import {
  createContext,
  type Dispatch,
  type ReactNode,
  type SetStateAction,
  useContext,
  useState,
} from 'react'
import type { Bill, Book } from '@/types/billing'
import useDialogState from '@/hooks/use-dialog-state'

type BillingDialogType =
  | 'add-book'
  | 'edit-book'
  | 'delete-book'
  | 'share-book'
  | 'add-bill'
  | 'delete-bill'

interface BillingContext {
  open: BillingDialogType | null
  setOpen: (str: BillingDialogType | null) => void
  currentBook: Book | null
  setCurrentBook: Dispatch<SetStateAction<Book | null>>
  currentBill: Bill | null
  setCurrentBill: Dispatch<SetStateAction<Bill | null>>
}

const BillingContext = createContext<BillingContext | null>(null)

export function BillingProvider({ children }: { children: ReactNode }) {
  const [open, setOpen] = useDialogState<BillingDialogType>(null)
  const [currentBook, setCurrentBook] = useState<Book | null>(null)
  const [currentBill, setCurrentBill] = useState<Bill | null>(null)

  return (
    <BillingContext
      value={{
        open,
        setOpen,
        currentBook,
        setCurrentBook,
        currentBill,
        setCurrentBill,
      }}
    >
      {children}
    </BillingContext>
  )
}

export const useBilling = () => {
  const context = useContext(BillingContext)
  if (!context) {
    throw new Error('useBilling has to be used within <BillingProvider>')
  }
  return context
}
