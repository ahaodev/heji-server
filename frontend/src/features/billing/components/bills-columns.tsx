import { type ColumnDef } from '@tanstack/react-table'
import type { Bill } from '@/types/billing'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import { LongText } from '@/components/long-text'
import { fenToYuan } from '../data/schema'

export const billColumns: ColumnDef<Bill>[] = [
  {
    accessorKey: 'time',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='日期' />
    ),
    cell: ({ row }) => (
      <span className='w-fit text-nowrap'>
        {row.original.time?.slice(0, 10)}
      </span>
    ),
    enableHiding: false,
  },
  {
    accessorKey: 'type',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='类型' />
    ),
    cell: ({ row }) => (
      <Badge
        variant='outline'
        className={cn(
          row.original.type === 1
            ? 'border-green-200 bg-green-100 text-green-800'
            : 'border-red-200 bg-red-100 text-red-800'
        )}
      >
        {row.original.type === 1 ? '收入' : '支出'}
      </Badge>
    ),
    filterFn: (row, id, value: string[]) =>
      Array.isArray(value) && value.includes(String(row.getValue(id))),
    enableSorting: false,
  },
  {
    accessorKey: 'category',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='分类' />
    ),
    cell: ({ row }) => <span>{row.original.category || '-'}</span>,
    enableSorting: false,
  },
  {
    accessorKey: 'money',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='金额' />
    ),
    cell: ({ row }) => {
      const { money, type } = row.original
      return (
        <span
          className={cn(
            'font-medium text-nowrap',
            type === 1 ? 'text-green-600' : 'text-red-600'
          )}
        >
          {type === 1 ? '+' : '-'}
          {fenToYuan(money).toFixed(2)}
        </span>
      )
    },
  },
  {
    accessorKey: 'remark',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='备注' />
    ),
    cell: ({ row }) => (
      <LongText className='max-w-48'>{row.original.remark || '-'}</LongText>
    ),
    enableSorting: false,
  },
  {
    accessorKey: 'crt_user_name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='创建人' />
    ),
    cell: ({ row }) => (
      <span className='w-fit text-nowrap'>
        {row.original.crt_user_name || row.original.crt_user}
      </span>
    ),
    enableSorting: false,
  },
]
