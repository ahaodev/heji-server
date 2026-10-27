import { useQuery } from '@tanstack/react-query'
import {
  createBill,
  createBook,
  deleteBill,
  deleteBook,
  getBill,
  getBillImages,
  getBills,
  getBooks,
  getCategories,
  joinBook,
  shareBook,
  updateBill,
  updateBook,
  uploadBillImage,
} from '@/services/billingApi'
import type {
  BillQueryParams,
  CreateBillRequest,
  CreateBookRequest,
  UpdateBillRequest,
  UpdateBookRequest,
} from '@/types/billing'
import { useCrudMutation } from '@/hooks/use-crud-mutation'

export const BOOKS_QUERY_KEY = 'books'
export const BILLS_QUERY_KEY = 'bills'
export const CATEGORIES_QUERY_KEY = 'categories'
export const BILL_IMAGES_QUERY_KEY = 'bill-images'

export function useBooks() {
  return useQuery({
    queryKey: [BOOKS_QUERY_KEY],
    queryFn: getBooks,
    staleTime: 5 * 60 * 1000,
  })
}

export function useBills(params: BillQueryParams) {
  return useQuery({
    queryKey: [BILLS_QUERY_KEY, params],
    queryFn: () => getBills(params),
    enabled: !!params.book_id,
    staleTime: 5 * 60 * 1000,
  })
}

export function useBill(id: string | undefined) {
  return useQuery({
    queryKey: [BILLS_QUERY_KEY, id],
    queryFn: () => getBill(id as string),
    enabled: !!id,
  })
}

export function useCategories(bookId: string | undefined) {
  return useQuery({
    queryKey: [CATEGORIES_QUERY_KEY, bookId],
    queryFn: () => getCategories(bookId as string),
    enabled: !!bookId,
    staleTime: 5 * 60 * 1000,
  })
}

export function useBillImages(billId: string | undefined) {
  return useQuery({
    queryKey: [BILL_IMAGES_QUERY_KEY, billId],
    queryFn: () => getBillImages(billId as string),
    enabled: !!billId,
  })
}

export function useCreateBook() {
  return useCrudMutation({
    mutationFn: (data: CreateBookRequest) => createBook(data),
    queryKeys: [[BOOKS_QUERY_KEY]],
    successMessage: '账本创建成功',
    errorMessage: '创建账本失败',
  })
}

export function useUpdateBook() {
  return useCrudMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateBookRequest }) =>
      updateBook(id, data),
    queryKeys: [[BOOKS_QUERY_KEY]],
    successMessage: '账本更新成功',
    errorMessage: '更新账本失败',
  })
}

export function useDeleteBook() {
  return useCrudMutation({
    mutationFn: deleteBook,
    queryKeys: [[BOOKS_QUERY_KEY]],
    successMessage: '账本删除成功',
    errorMessage: '删除账本失败',
  })
}

export function useJoinBook() {
  return useCrudMutation({
    mutationFn: joinBook,
    queryKeys: [[BOOKS_QUERY_KEY]],
    successMessage: '已加入账本',
    errorMessage: '加入账本失败',
  })
}

export function useShareBook() {
  return useCrudMutation({
    mutationFn: shareBook,
    queryKeys: [],
    successMessage: '分享码已生成',
    errorMessage: '生成分享码失败',
  })
}

export function useCreateBill() {
  return useCrudMutation({
    mutationFn: (data: CreateBillRequest) => createBill(data),
    queryKeys: [[BILLS_QUERY_KEY]],
    successMessage: '账单创建成功',
    errorMessage: '创建账单失败',
  })
}

export function useUpdateBill() {
  return useCrudMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateBillRequest }) =>
      updateBill(id, data),
    queryKeys: [[BILLS_QUERY_KEY]],
    successMessage: '账单更新成功',
    errorMessage: '更新账单失败',
  })
}

export function useDeleteBill() {
  return useCrudMutation({
    mutationFn: deleteBill,
    queryKeys: [[BILLS_QUERY_KEY]],
    successMessage: '账单删除成功',
    errorMessage: '删除账单失败',
  })
}

export function useUploadBillImage() {
  return useCrudMutation({
    mutationFn: ({ billId, file }: { billId: string; file: File }) =>
      uploadBillImage(billId, file),
    queryKeys: (_, { billId }) => [[BILL_IMAGES_QUERY_KEY, billId]],
    successMessage: '图片上传成功',
    errorMessage: '图片上传失败',
  })
}
