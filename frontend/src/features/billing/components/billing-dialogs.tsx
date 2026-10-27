import type { Bill, Book } from '@/types/billing'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  useDeleteBill,
  useDeleteBook,
  useShareBook,
} from '../hooks/use-billing'
import { BillActionDialog } from './bill-action-dialog'
import { useBilling } from './billing-provider'
import { BookActionDialog } from './book-action-dialog'

interface BillingDialogsProps {
  /** 账本 ID；账单页传入后才渲染账单相关对话框。 */
  bookId?: string
  categories?: { _id: string; name: string }[]
}

export function BillingDialogs({
  bookId,
  categories = [],
}: BillingDialogsProps) {
  const {
    open,
    setOpen,
    currentBook,
    setCurrentBook,
    currentBill,
    setCurrentBill,
  } = useBilling()

  return (
    <>
      <BookActionDialog
        key='book-add'
        open={open === 'add-book'}
        onOpenChange={() => setOpen('add-book')}
      />

      {currentBook && (
        <>
          <BookActionDialog
            key={`book-edit-${currentBook._id}`}
            open={open === 'edit-book'}
            onOpenChange={() => {
              setOpen('edit-book')
              setTimeout(() => setCurrentBook(null), 500)
            }}
            currentBook={currentBook}
          />
          <BookDeleteDialog
            key={`book-delete-${currentBook._id}`}
            open={open === 'delete-book'}
            onOpenChange={() => {
              setOpen('delete-book')
              setTimeout(() => setCurrentBook(null), 500)
            }}
            book={currentBook}
          />
          <BookShareDialog
            key={`book-share-${currentBook._id}`}
            open={open === 'share-book'}
            onOpenChange={() => {
              setOpen('share-book')
              setTimeout(() => setCurrentBook(null), 500)
            }}
            book={currentBook}
          />
        </>
      )}

      {bookId && (
        <BillActionDialog
          key='bill-add'
          bookId={bookId}
          categories={categories}
          open={open === 'add-bill' && !currentBill}
          onOpenChange={() => setOpen('add-bill')}
        />
      )}

      {bookId && currentBill && (
        <>
          <BillActionDialog
            key={`bill-edit-${currentBill._id}`}
            bookId={bookId}
            categories={categories}
            currentBill={currentBill}
            open={open === 'add-bill'}
            onOpenChange={() => {
              setOpen('add-bill')
              setTimeout(() => setCurrentBill(null), 500)
            }}
          />
          <BillDeleteDialog
            key={`bill-delete-${currentBill._id}`}
            bill={currentBill}
            open={open === 'delete-bill'}
            onOpenChange={() => {
              setOpen('delete-bill')
              setTimeout(() => setCurrentBill(null), 500)
            }}
          />
        </>
      )}
    </>
  )
}

interface BookDeleteDialogProps {
  book: Book | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

function BookDeleteDialog({ book, open, onOpenChange }: BookDeleteDialogProps) {
  const deleteBook = useDeleteBook()
  if (!book) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>删除账本</DialogTitle>
          <DialogDescription>
            确定要删除账本 <strong>“{book.name}”</strong> 吗？
            账本下的账单将一并不可见，此操作不可撤销。
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button
            variant='outline'
            onClick={() => onOpenChange(false)}
            disabled={deleteBook.isPending}
          >
            取消
          </Button>
          <Button
            variant='destructive'
            onClick={() =>
              deleteBook.mutate(book._id, {
                onSuccess: () => onOpenChange(false),
              })
            }
            disabled={deleteBook.isPending}
          >
            {deleteBook.isPending ? '删除中...' : '删除账本'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

interface BillDeleteDialogProps {
  bill: Bill | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

function BillDeleteDialog({ bill, open, onOpenChange }: BillDeleteDialogProps) {
  const deleteBill = useDeleteBill()
  if (!bill) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>删除账单</DialogTitle>
          <DialogDescription>
            确定要删除 {bill.time?.slice(0, 10)}{' '}
            的这条账单吗？删除后仍会同步给其他成员。
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button
            variant='outline'
            onClick={() => onOpenChange(false)}
            disabled={deleteBill.isPending}
          >
            取消
          </Button>
          <Button
            variant='destructive'
            onClick={() =>
              deleteBill.mutate(bill._id, {
                onSuccess: () => onOpenChange(false),
              })
            }
            disabled={deleteBill.isPending}
          >
            {deleteBill.isPending ? '删除中...' : '删除账单'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

interface BookShareDialogProps {
  book: Book | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

function BookShareDialog({ book, open, onOpenChange }: BookShareDialogProps) {
  const shareBook = useShareBook()

  return (
    <Dialog
      open={open}
      onOpenChange={(state) => {
        if (state && !shareBook.data && book) shareBook.mutate(book._id)
        onOpenChange(state)
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>分享账本</DialogTitle>
          <DialogDescription>
            把下面的分享码发给对方，对方在“加入账本”中填写即可成为成员。
          </DialogDescription>
        </DialogHeader>
        <div className='bg-muted/40 flex items-center gap-2 rounded-md border px-3 py-2 font-mono text-sm'>
          <span className='flex-1 truncate'>
            {shareBook.data ?? (shareBook.isPending ? '生成中...' : '')}
          </span>
          {shareBook.data && (
            <Button
              variant='ghost'
              size='sm'
              onClick={() => navigator.clipboard.writeText(shareBook.data)}
            >
              复制
            </Button>
          )}
        </div>
        <DialogFooter>
          <Button onClick={() => onOpenChange(false)}>关闭</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
