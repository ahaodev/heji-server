import { useNavigate } from '@tanstack/react-router'
import {
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { Plus } from 'lucide-react'
import { PERMISSIONS } from '@/constants/permissions'
import { usePermission } from '@/hooks/usePermission'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Main } from '@/components/layout/main'
import { PageHeader } from '@/components/layout/page-header'
import { BillingDialogs } from './components/billing-dialogs'
import { BillingProvider, useBilling } from './components/billing-provider'
import { createBookColumns } from './components/books-columns'
import { useBooks } from './hooks/use-billing'

export function Books() {
  return (
    <BillingProvider>
      <BooksContent />
      <BillingDialogs />
    </BillingProvider>
  )
}

function BooksContent() {
  const { setOpen, setCurrentBook } = useBilling()
  const { hasPermission } = usePermission()
  const { data: books, isLoading, error } = useBooks()
  const navigate = useNavigate()

  const columns = createBookColumns({
    hasPermission,
    onEdit: (book) => {
      setCurrentBook(book)
      setOpen('edit-book')
    },
    onDelete: (book) => {
      setCurrentBook(book)
      setOpen('delete-book')
    },
    onShare: (book) => {
      setCurrentBook(book)
      setOpen('share-book')
    },
  })

  const table = useReactTable({
    data: books ?? [],
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
  })

  return (
    <>
      <PageHeader fixed />

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h2 className='text-2xl font-bold tracking-tight'>账本管理</h2>
            <p className='text-muted-foreground'>
              浏览您的所有账本，点击行可查看账单明细。
            </p>
          </div>
          {hasPermission(PERMISSIONS.BILLING.BOOK.WRITE) && (
            <Button
              onClick={() => {
                setCurrentBook(null)
                setOpen('add-book')
              }}
            >
              新增账本
              <Plus className='ms-1 h-4 w-4' />
            </Button>
          )}
        </div>

        {isLoading ? (
          <div className='space-y-2'>
            {Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} className='h-16 w-full' />
            ))}
          </div>
        ) : error ? (
          <div className='text-muted-foreground flex h-32 items-center justify-center'>
            加载账本数据失败，请重试
          </div>
        ) : (
          <div className='overflow-hidden rounded-md border'>
            <Table>
              <TableHeader>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <TableHead key={header.id}>
                        {header.isPlaceholder
                          ? null
                          : flexRender(
                              header.column.columnDef.header,
                              header.getContext()
                            )}
                      </TableHead>
                    ))}
                  </TableRow>
                ))}
              </TableHeader>
              <TableBody>
                {table.getRowModel().rows.length ? (
                  table.getRowModel().rows.map((row) => (
                    <TableRow
                      key={row.id}
                      className='cursor-pointer'
                      onClick={(event) => {
                        // 操作列下拉菜单不应触发进入账本
                        if ((event.target as HTMLElement).closest('button'))
                          return
                        navigate({
                          to: '/books/$bookId',
                          params: { bookId: row.original._id },
                        })
                      }}
                    >
                      {row.getVisibleCells().map((cell) => (
                        <TableCell key={cell.id}>
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext()
                          )}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell
                      colSpan={columns.length}
                      className='h-24 text-center'
                    >
                      还没有账本，先创建一个。
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        )}
      </Main>
    </>
  )
}
