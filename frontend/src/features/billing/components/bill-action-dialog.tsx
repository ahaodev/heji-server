import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import type { Bill } from '@/types/billing'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import {
  billFormSchema,
  fenToYuan,
  yuanToFen,
  type BillFormData,
} from '../data/schema'
import { useCreateBill, useUpdateBill } from '../hooks/use-billing'

interface BillActionDialogProps {
  bookId: string
  currentBill?: Bill | null
  categories: { _id: string; name: string }[]
  open: boolean
  onOpenChange: (open: boolean) => void
}

const today = () => new Date().toISOString().slice(0, 10)

export function BillActionDialog({
  bookId,
  currentBill,
  categories,
  open,
  onOpenChange,
}: BillActionDialogProps) {
  const isEdit = !!currentBill
  const createBill = useCreateBill()
  const updateBill = useUpdateBill()
  const isSubmitting = createBill.isPending || updateBill.isPending

  const form = useForm<BillFormData>({
    resolver: zodResolver(billFormSchema),
    defaultValues: {
      money: 0,
      type: -1,
      category: '',
      time: today(),
      remark: '',
    },
  })

  useEffect(() => {
    if (!open) return
    form.reset({
      money: currentBill ? fenToYuan(currentBill.money) : 0,
      type: (currentBill?.type ?? -1) as 1 | -1,
      category: currentBill?.category ?? '',
      time: currentBill?.time?.slice(0, 10) ?? today(),
      remark: currentBill?.remark ?? '',
    })
  }, [open, currentBill, form])

  const onSubmit = (values: BillFormData) => {
    const money = yuanToFen(values.money)
    if (isEdit && currentBill) {
      updateBill.mutate(
        {
          id: currentBill._id,
          data: { ...values, money },
        },
        { onSuccess: () => onOpenChange(false) }
      )
      return
    }
    createBill.mutate(
      { ...values, money, book_id: bookId },
      { onSuccess: () => onOpenChange(false) }
    )
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
          <DialogTitle>{isEdit ? '编辑账单' : '新增账单'}</DialogTitle>
          <DialogDescription>
            金额以元为单位录入，保存时会换算为分。
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form
            id='bill-form'
            onSubmit={form.handleSubmit(onSubmit)}
            className='space-y-4'
          >
            <div className='grid grid-cols-2 gap-3'>
              <FormField
                control={form.control}
                name='money'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>金额（元）</FormLabel>
                    <FormControl>
                      <Input
                        type='number'
                        step='0.01'
                        min='0'
                        value={field.value}
                        onChange={(event) =>
                          field.onChange(
                            Number.isNaN(event.target.valueAsNumber)
                              ? 0
                              : event.target.valueAsNumber
                          )
                        }
                        onBlur={field.onBlur}
                        name={field.name}
                        ref={field.ref}
                      />
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
                    <FormLabel>类型</FormLabel>
                    <Select
                      value={String(field.value)}
                      onValueChange={(value) =>
                        field.onChange(value === '1' ? 1 : -1)
                      }
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value='-1'>支出</SelectItem>
                        <SelectItem value='1'>收入</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className='grid grid-cols-2 gap-3'>
              <FormField
                control={form.control}
                name='category'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>分类</FormLabel>
                    <Select
                      value={field.value || '__none__'}
                      onValueChange={(value) =>
                        field.onChange(value === '__none__' ? '' : value)
                      }
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder='未分类' />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value='__none__'>未分类</SelectItem>
                        {categories.map((category) => (
                          <SelectItem key={category._id} value={category.name}>
                            {category.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='time'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>日期</FormLabel>
                    <FormControl>
                      <Input type='date' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name='remark'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>备注</FormLabel>
                  <FormControl>
                    <Textarea rows={2} placeholder='可选' {...field} />
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
          <Button type='submit' form='bill-form' disabled={isSubmitting}>
            {isSubmitting ? '保存中...' : '保存'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
