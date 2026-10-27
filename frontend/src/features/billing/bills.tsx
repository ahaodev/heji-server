import { useEffect, useMemo, useState } from 'react'
import { Link, getRouteApi } from '@tanstack/react-router'
import {
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
  type SortingState,
} from '@tanstack/react-table'
import { exportBills } from '@/services/billingApi'
import type { Bill } from '@/types/billing'
import { ChevronLeft, Download, Pencil, Plus, Trash2 } from 'lucide-react'
import { PERMISSIONS } from '@/constants/permissions'
import { type NavigateFn, useTableUrlState } from '@/hooks/use-table-url-state'
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
import { DataTablePagination, DataTableToolbar } from '@/components/data-table'
import { Main } from '@/components/layout/main'
import { PageHeader } from '@/components/layout/page-header'
import { BillingDialogs } from './components/billing-dialogs'
import { BillingProvider, useBilling } from './components/billing-provider'
import { billColumns } from './components/bills-columns'
import { useBills, useBooks, useCategories } from './hooks/use-billing'

const route = getRouteApi('/_authenticated/books/$bookId')

export function BookBills() {
  const { bookId } = route.useParams()
  const { data: categories } = useCategories(bookId)
  const categoryOptions = useMemo(
    () => (categories ?? []).map((c) => ({ _id: c._id, name: c.name })),
    [categories]
  )

  return (
    <BillingProvider>
      <BookBillsContent bookId={bookId} />
      <BillingDialogs bookId={bookId} categories={categoryOptions} />
    </BillingProvider>
  )
}

interface BookBillsContentProps {
  bookId: string
}

function BookBillsContent({ bookId }: BookBillsContentProps) {
  const search = route.useSearch()
  const navigate = route.useNavigate()
  const { hasPermission } = usePermission()
  const { setOpen, setCurrentBill, currentBook } = useBilling()

  const { data: books } = useBooks()
  const bookName =
    books?.find((book) => book._id === bookId)?.name ??
    currentBook?.name ??
    '账本'

  const queryParams = {
    book_id: bookId,
    page: search.page || 1,
    page_size: search.pageSize || 10,
  }
  const { data: billsData, isLoading, error } = useBills(queryParams)

  const navigateWrapper: NavigateFn = ({ search: searchUpdate, replace }) => {
    navigate({ search: searchUpdate, replace })
  }

  return (
    <>
      <PageHeader fixed />

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div className='flex items-center gap-2'>
            <Button variant='ghost' size='icon' className='size-8' asChild>
              <Link to='/books'>
                <ChevronLeft className='size-4' />
              </Link>
            </Button>
            <div>
              <h2 className='text-2xl font-bold tracking-tight'>{bookName}</h2>
              <p className='text-muted-foreground'>该账本下的全部账单记录。</p>
            </div>
          </div>
          <div className='flex gap-2'>
            <Button
              variant='outline'
              onClick={async () => {
                const blob = await exportBills(bookId)
                const url = URL.createObjectURL(blob)
                const anchor = document.createElement('a')
                anchor.href = url
                anchor.download = `${bookName}-账单.csv`
                anchor.click()
                URL.revokeObjectURL(url)
              }}
            >
              <Download className='me-1 h-4 w-4' />
              导出
            </Button>
            {hasPermission(PERMISSIONS.BILLING.BILL.WRITE) && (
              <Button
                onClick={() => {
                  setCurrentBill(null)
                  setOpen('add-bill')
                }}
              >
                新增账单
                <Plus className='ms-1 h-4 w-4' />
              </Button>
            )}
          </div>
        </div>

        {isLoading ? (
          <div className='space-y-2'>
            {Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} className='h-16 w-full' />
            ))}
          </div>
        ) : error ? (
          <div className='text-muted-foreground flex h-32 items-center justify-center'>
            加载账单数据失败，请重试
          </div>
        ) : (
          <BillsTable
            data={billsData?.list ?? []}
            search={search}
            navigate={navigateWrapper}
            totalCount={billsData?.total ?? 0}
            canWrite={hasPermission(PERMISSIONS.BILLING.BILL.WRITE)}
            onEdit={(bill) => {
              setCurrentBill(bill)
              setOpen('add-bill')
            }}
            onDelete={(bill) => {
              setCurrentBill(bill)
              setOpen('delete-bill')
            }}
          />
        )}
      </Main>
    </>
  )
}

interface BillsTableProps {
  data: Bill[]
  search: Record<string, unknown>
  navigate: NavigateFn
  totalCount: number
  canWrite: boolean
  onEdit: (bill: Bill) => void
  onDelete: (bill: Bill) => void
}

function BillsTable({
  data,
  search,
  navigate,
  totalCount,
  canWrite,
  onEdit,
  onDelete,
}: BillsTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const {
    columnFilters,
    onColumnFiltersChange,
    pagination,
    onPaginationChange,
    ensurePageInRange,
  } = useTableUrlState({
    search,
    navigate,
    pagination: { defaultPage: 1, defaultPageSize: 10 },
    globalFilter: { enabled: false },
    columnFilters: [{ columnId: 'type', searchKey: 'type', type: 'array' }],
  })

  const table = useReactTable({
    data,
    columns: billColumns,
    state: { sorting, pagination, columnFilters },
    manualPagination: true,
    manualFiltering: true,
    pageCount: Math.ceil(totalCount / (pagination.pageSize || 10)),
    onPaginationChange,
    onColumnFiltersChange,
    onSortingChange: setSorting,
    getPaginationRowModel: getPaginationRowModel(),
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
  })

  useEffect(() => {
    ensurePageInRange(table.getPageCount())
  }, [table, ensurePageInRange])

  return (
    <div className='space-y-4'>
      <DataTableToolbar
        table={table}
        filters={[
          {
            columnId: 'type',
            title: '类型',
            options: [
              { label: '收入', value: '1' },
              { label: '支出', value: '-1' },
            ],
          },
        ]}
      />
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
                {canWrite && <TableHead className='w-24'>操作</TableHead>}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                  {canWrite && (
                    <TableCell>
                      <div className='flex gap-1'>
                        <Button
                          variant='ghost'
                          size='icon'
                          className='size-8'
                          onClick={() => onEdit(row.original)}
                        >
                          <Pencil className='h-4 w-4' />
                        </Button>
                        <Button
                          variant='ghost'
                          size='icon'
                          className='size-8'
                          onClick={() => onDelete(row.original)}
                        >
                          <Trash2 className='h-4 w-4' />
                        </Button>
                      </div>
                    </TableCell>
                  )}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={billColumns.length + (canWrite ? 1 : 0)}
                  className='h-24 text-center'
                >
                  没有账单记录。
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <DataTablePagination table={table} />
    </div>
  )
}
