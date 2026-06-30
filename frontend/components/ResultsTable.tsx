'use client'

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Download } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface Product {
  product_name: string
  price: number
  currency: string
}

interface ResultsTableProps {
  products: Product[]
  onExport: () => void
}

export function ResultsTable({ products, onExport }: ResultsTableProps) {
  if (products.length === 0) {
    return null
  }

  return (
    <Card className="w-full max-w-4xl mx-auto">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Результаты парсинга</CardTitle>
          <Button onClick={onExport} size="sm">
            <Download className="w-4 h-4 mr-2" />
            Экспорт CSV
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Товар</TableHead>
                <TableHead className="text-right">Цена</TableHead>
                <TableHead>Валюта</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {products.map((product, index) => (
                <TableRow key={index}>
                  <TableCell className="font-medium">{product.product_name}</TableCell>
                  <TableCell className="text-right">{product.price.toFixed(2)}</TableCell>
                  <TableCell>{product.currency}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  )
}
