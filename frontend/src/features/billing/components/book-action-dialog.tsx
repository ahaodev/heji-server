import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import type { Book } from '@/types/billing'
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
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { bookFormSchema, type BookFormData } from '../data/schema'
import { useCreateBook, useUpdateBook } from '../hooks/use-billing'

interface BookActionDialogProps {
  currentBook?: Book | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function BookActionDialog({
  currentBook,
  open,
  onOpenChange,
}: BookActionDialogProps) {
  const isEdit = !!currentBook
  const createBook = useCreateBook()
  const updateBook = useUpdateBook()
  const isSubmitting = createBook.isPending || updateBook.isPending

  const form = useForm<BookFormData>({
    resolver: zodResolver(bookFormSchema),
    defaultValues: { name: '', type: '', banner: '' },
  })

  useEffect(() => {
    if (!open) return
    form.reset({
      name: currentBook?.name ?? '',
      type: currentBook?.type ?? '',
      banner: currentBook?.banner ?? '',
    })
  }, [open, currentBook, form])

  const onSubmit = (values: BookFormData) => {
    if (isEdit && currentBook) {
      updateBook.mutate(
        { id: currentBook._id, data: values },
        { onSuccess: () => onOpenChange(false) }
      )
      return
    }
    createBook.mutate(values, { onSuccess: () => onOpenChange(false) })
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(state) => {
        if (!state) form.reset()
        onOpenChange(state)
      }}
    >
      <DialogContent className='sm:max-w-md'>
        <DialogHeader className='text-start'>
          <DialogTitle>{isEdit ? '编辑账本' : '新增账本'}</DialogTitle>
          <DialogDescription>
            {isEdit ? '修改账本信息。' : '创建一个新账本。'}
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form
            id='book-form'
            onSubmit={form.handleSubmit(onSubmit)}
            className='space-y-4'
          >
            <FormField
              control={form.control}
              name='name'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>账本名称</FormLabel>
                  <FormControl>
                    <Input placeholder='例如：家庭账本' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='type'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>类型（可选）</FormLabel>
                  <FormControl>
                    <Input placeholder='例如：日常' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='banner'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>封面 URL（可选）</FormLabel>
                  <FormControl>
                    <Input placeholder='https://...' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </form>
        </Form>

        <DialogFooter>
          <Button
            variant='outline'
            onClick={() => onOpenChange(false)}
            disabled={isSubmitting}
          >
            取消
          </Button>
          <Button type='submit' form='book-form' disabled={isSubmitting}>
            {isSubmitting ? '保存中...' : '保存'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
