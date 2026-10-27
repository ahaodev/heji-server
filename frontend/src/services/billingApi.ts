import type { PagedResult } from '@/types/api'
import type {
  Bill,
  BillImage,
  BillQueryParams,
  Book,
  Category,
  CreateBillRequest,
  CreateBookRequest,
  CreateCategoryRequest,
  SyncChangesResponse,
  UpdateBillRequest,
  UpdateBookRequest,
} from '@/types/billing'
import { buildSearchParams } from '@/lib/query-params'
import { apiClient } from './config'

// ---- 账本 ----

// GET /book - 当前用户可访问的账本
export async function getBooks(): Promise<Book[]> {
  const response = await apiClient.get('/api/v1/book')
  return response.data.data ?? []
}

// GET /book/:id - 账本详情
export async function getBook(id: string): Promise<Book> {
  const response = await apiClient.get(`/api/v1/book/${id}`)
  return response.data.data
}

// POST /book - 创建账本
export async function createBook(data: CreateBookRequest): Promise<Book> {
  const response = await apiClient.post('/api/v1/book', data)
  return response.data.data
}

// PUT /book/:id - 更新账本
export async function updateBook(
  id: string,
  data: UpdateBookRequest
): Promise<Book> {
  const response = await apiClient.put(`/api/v1/book/${id}`, data)
  return response.data.data
}

// DELETE /book/:id - 删除账本
export async function deleteBook(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/book/${id}`)
}

// POST /book/:id/join - 用分享码加入账本
export async function joinBook(id: string): Promise<void> {
  await apiClient.post(`/api/v1/book/${id}/join`)
}

// POST /book/:id/share - 生成分享码
export async function shareBook(id: string): Promise<string> {
  const response = await apiClient.post(`/api/v1/book/${id}/share`)
  return response.data.data
}

// ---- 账单 ----

// GET /bill - 账单分页列表
export async function getBills(
  params: BillQueryParams
): Promise<PagedResult<Bill>> {
  const searchParams = buildSearchParams(params)
  const response = await apiClient.get(`/api/v1/bill?${searchParams}`)
  return response.data.data
}

// GET /bill/:id - 账单详情
export async function getBill(id: string): Promise<Bill> {
  const response = await apiClient.get(`/api/v1/bill/${id}`)
  return response.data.data
}

// POST /bill - 创建账单
export async function createBill(data: CreateBillRequest): Promise<Bill> {
  const response = await apiClient.post('/api/v1/bill', data)
  return response.data.data
}

// PUT /bill/:id - 更新账单
export async function updateBill(
  id: string,
  data: UpdateBillRequest
): Promise<Bill> {
  const response = await apiClient.put(`/api/v1/bill/${id}`, data)
  return response.data.data
}

// DELETE /bill/:id - 删除账单
export async function deleteBill(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/bill/${id}`)
}

// GET /bill/export - 导出 CSV（浏览器直接下载，带 Bearer 头会触发刷新逻辑，因此用 blob 请求）
export async function exportBills(
  bookId: string,
  year?: string,
  month?: string
): Promise<Blob> {
  const response = await apiClient.get('/api/v1/bill/export', {
    params: { book_id: bookId, year, month },
    responseType: 'blob',
  })
  return response.data
}

// ---- 分类 ----

// GET /category - 账本下的分类
export async function getCategories(bookId: string): Promise<Category[]> {
  const response = await apiClient.get('/api/v1/category', {
    params: { book_id: bookId },
  })
  return response.data.data ?? []
}

// POST /category - 创建分类
export async function createCategory(
  data: CreateCategoryRequest
): Promise<Category> {
  const response = await apiClient.post('/api/v1/category', data)
  return response.data.data
}

// DELETE /category/:id - 删除分类
export async function deleteCategory(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/category/${id}`)
}

// ---- 账单图片 ----

// GET /image/list - 账单的图片列表
export async function getBillImages(billId: string): Promise<BillImage[]> {
  const response = await apiClient.get('/api/v1/image/list', {
    params: { bill_id: billId },
  })
  return response.data.data ?? []
}

// POST /image/upload - 上传账单图片
export async function uploadBillImage(
  billId: string,
  file: File
): Promise<BillImage> {
  const form = new FormData()
  form.append('file', file)
  const response = await apiClient.post('/api/v1/image/upload', form, {
    params: { billId },
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return response.data.data
}

// 图片内容需要 Bearer 头，<img src> 无法携带，改为 blob URL
export async function getBillImageObjectURL(imageId: string): Promise<string> {
  const response = await apiClient.get(`/api/v1/image/${imageId}`, {
    responseType: 'blob',
  })
  return URL.createObjectURL(response.data)
}

// ---- 增量同步 ----

// GET /sync/changes - 拉取 since 之后的变更
export async function getSyncChanges(
  since: number,
  limit?: number
): Promise<SyncChangesResponse> {
  const response = await apiClient.get('/api/v1/sync/changes', {
    params: { since, limit },
  })
  return response.data.data
}
