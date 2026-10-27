import { type ColumnDef } from '@tanstack/react-table'
import type { Book } from '@/types/billing'
import { MoreHorizontal, Pencil, Share2, Trash2 } from 'lucide-react'
import { PERMISSIONS } from '@/constants/permissions'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { DataTableColumnHeader } from '@/components/data-table'
import { LongText } from '@/components/long-text'

interface BookColumnsProps {
  hasPermission: (permission: string) => boolean
  onEdit: (book: Book) => void
  onDelete: (book: Book) => void
  onShare: (book: Book) => void
}

export function createBookColumns({
  hasPermission,
  onEdit,
  onDelete,
  onShare,
}: BookColumnsProps): ColumnDef<Book>[] {
  const canWrite = hasPermission(PERMISSIONS.BILLING.BOOK.WRITE)

  return [
    {
      accessorKey: 'name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='账本名称' />
      ),
      cell: ({ row }) => (
        <LongText className='max-w-48'>{row.original.name}</LongText>
      ),
      enableHiding: false,
    },
    {
      accessorKey: 'type',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='类型' />
      ),
      cell: ({ row }) => (
        <span className='text-muted-foreground'>
          {row.original.type || '-'}
        </span>
      ),
      enableSorting: false,
    },
    {
      accessorKey: 'is_initial',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='默认' />
      ),
      cell: ({ row }) =>
        row.original.is_initial ? (
          <Badge variant='secondary'>默认</Badge>
        ) : (
          <span className='text-muted-foreground'>-</span>
        ),
      enableSorting: false,
    },
    {
      id: 'members',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='成员数' />
      ),
      cell: ({ row }) => <span>{row.original.members?.length ?? 0}</span>,
      enableSorting: false,
    },
    {
      accessorKey: 'upd_time',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='更新时间' />
      ),
      cell: ({ row }) => (
        <span className='text-muted-foreground'>
          {row.original.upd_time?.replace('T', ' ').slice(0, 19)}
        </span>
      ),
    },
    {
      id: 'actions',
      cell: ({ row }) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant='ghost' className='h-8 w-8 p-0'>
              <MoreHorizontal className='h-4 w-4' />
              <span className='sr-only'>打开菜单</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align='end'>
            <DropdownMenuItem onClick={() => onShare(row.original)}>
              <Share2 className='me-2 h-4 w-4' />
              分享账本
            </DropdownMenuItem>
            {canWrite && (
              <DropdownMenuItem onClick={() => onEdit(row.original)}>
                <Pencil className='me-2 h-4 w-4' />
                编辑
              </DropdownMenuItem>
            )}
            {canWrite && (
              <DropdownMenuItem
                variant='destructive'
                onClick={() => onDelete(row.original)}
              >
                <Trash2 className='me-2 h-4 w-4' />
                删除
              </DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      ),
      enableHiding: false,
    },
  ]
}
