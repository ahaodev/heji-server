import { z } from 'zod'

// 金额以“元”输入，提交时换算为分（后端 money 单位为分）
export const bookFormSchema = z.object({
  name: z.string().min(1, '账本名称不能为空').max(100, '最多 100 个字符'),
  type: z.string().max(50, '最多 50 个字符').optional(),
  banner: z.string().max(255, '最多 255 个字符').optional(),
})

export type BookFormData = z.infer<typeof bookFormSchema>

export const billFormSchema = z.object({
  money: z.number().min(0.01, '金额必须大于 0'),
  type: z.union([z.literal(1), z.literal(-1)]),
  category: z.string().max(100, '最多 100 个字符').optional(),
  time: z.string().min(1, '请选择账单日期'),
  remark: z.string().max(500, '最多 500 个字符').optional(),
})

export type BillFormData = z.infer<typeof billFormSchema>

// 元 ↔ 分：0.01 精度下用四舍五入避免浮点误差放大
export const yuanToFen = (yuan: number): number => Math.round(yuan * 100)
export const fenToYuan = (fen: number): number => Number((fen / 100).toFixed(2))
