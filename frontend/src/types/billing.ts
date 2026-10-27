// 合記记账域：账本 / 账单 / 分类 / 账单图片。
// 字段与后端 domain.Book / Bill / Category / BillImage 的 JSON 契约一致，
// 金额单位是“分”，时间字段以客户端可读的字符串保留。

export interface Book {
  _id: string
  name: string
  type?: string
  banner?: string
  crt_user_id: string
  is_initial: boolean
  members: string[]
  deleted_at?: string
  crt_time: string
  upd_time: string
}

export interface CreateBookRequest {
  name: string
  type?: string
  banner?: string
  is_initial?: boolean
}

export interface UpdateBookRequest {
  name?: string
  type?: string
  banner?: string
}

export type BillType = 1 | -1

export interface Bill {
  _id: string
  book_id: string
  money: number // 分
  type: BillType
  category?: string
  crt_user: string
  crt_user_name?: string
  time: string
  remark?: string
  images?: string[]
  deleted_at?: string
  crt_time: string
  upd_time: string
}

export interface CreateBillRequest {
  _id?: string
  book_id: string
  money: number
  type: BillType
  category?: string
  time: string
  remark?: string
  images?: string[]
}

export interface UpdateBillRequest {
  money?: number
  type?: BillType
  category?: string
  time?: string
  remark?: string
  images?: string[]
}

export interface BillQueryParams {
  book_id: string
  type?: BillType
  category?: string
  time_from?: string
  time_to?: string
  page?: number
  page_size?: number
}

export interface Category {
  _id: string
  book_id: string
  name: string
  type: BillType
  level: number
  index: number
  parentId?: string
  create_user?: string
}

export interface CreateCategoryRequest {
  book_id: string
  name: string
  type: BillType
  level?: number
  index?: number
  parentId?: string
}

export interface BillImage {
  _id: string
  bill_id: string
  key: string
  ext?: string
  length?: number
  md5?: string
}

export interface SyncChangesResponse {
  books: Book[]
  bills: Bill[]
  has_more: boolean
  next_since: number
}
