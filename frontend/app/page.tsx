'use client'

import { useState, useCallback } from 'react'
import { useDropzone } from 'react-dropzone'
import { 
  Upload, File, X, Loader2, Download, CheckCircle, 
  AlertCircle, Trash2, Database, Brain, FileSpreadsheet,
  Clock, Layers, Sparkles
} from 'lucide-react'

interface Product {
  product_name: string
  price: number
  currency: string
}

interface FileStatus {
  name: string
  status: 'pending' | 'processing' | 'done' | 'error'
  productsFound: number
  size: number
}

const formatNumber = (num: number): string => {
  return new Intl.NumberFormat('ru-RU').format(num)
}

const formatFileSize = (bytes: number): string => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

export default function Home() {
  const [files, setFiles] = useState<File[]>([])
  const [fileStatuses, setFileStatuses] = useState<FileStatus[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [progress, setProgress] = useState(0)
  const [statusMessage, setStatusMessage] = useState('')
  const [currentFile, setCurrentFile] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [processingTime, setProcessingTime] = useState(0)
  const [totalFiles, setTotalFiles] = useState(0)
  const [processedFiles, setProcessedFiles] = useState(0)

  const onDrop = useCallback((acceptedFiles: File[]) => {
    const newFiles = [...files, ...acceptedFiles]
    setFiles(newFiles)
    setFileStatuses(newFiles.map(f => ({
      name: f.name,
      status: 'pending',
      productsFound: 0,
      size: f.size
    })))
    setError(null)
  }, [files])

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': ['.xlsx'],
      'application/vnd.ms-excel': ['.xls'],
      'application/vnd.ms-excel.sheet.macroEnabled.12': ['.xlsm'],
      'text/csv': ['.csv'],
      'application/pdf': ['.pdf'],
      'text/plain': ['.txt'],
    },
    multiple: true,
  })

  const removeFile = (index: number) => {
    setFiles(prev => prev.filter((_, i) => i !== index))
    setFileStatuses(prev => prev.filter((_, i) => i !== index))
  }

  const clearAll = () => {
    setFiles([])
    setFileStatuses([])
    setProducts([])
    setError(null)
    setProgress(0)
    setStatusMessage('')
    setCurrentFile('')
  }

  const handleUpload = async () => {
    if (files.length === 0) {
      setError('Пожалуйста, выберите файлы для загрузки')
      return
    }

    setIsLoading(true)
    setProgress(0)
    setStatusMessage('Подготовка файлов...')
    setError(null)
    setProducts([])
    setProcessingTime(0)
    setTotalFiles(files.length)
    setProcessedFiles(0)

    setFileStatuses(files.map(f => ({
      name: f.name,
      status: 'pending',
      productsFound: 0,
      size: f.size
    })))

    const startTime = Date.now()
    let totalProducts: Product[] = []

    for (let i = 0; i < files.length; i++) {
      const file = files[i]
      const fileName = file.name
      setCurrentFile(fileName)
      
      setFileStatuses(prev => prev.map((fs, idx) => 
        idx === i ? { ...fs, status: 'processing' } : fs
      ))
      
      setStatusMessage(`Обработка файла ${i + 1} из ${files.length}: ${fileName}`)
      setProgress(Math.round((i / files.length) * 100))
      
      const formData = new FormData()
      formData.append('file', file)

      try {
        const controller = new AbortController()
        const timeoutId = setTimeout(() => controller.abort(), 600000)

        setStatusMessage(`AI анализ файла ${i + 1} из ${files.length}...`)
        
        const response = await fetch('http://localhost:8080/api/parse-file', {
          method: 'POST',
          body: formData,
          signal: controller.signal,
        })

        clearTimeout(timeoutId)

        if (!response.ok) {
          const err = await response.json()
          throw new Error(`Ошибка в файле ${file.name}: ${err.error || 'Неизвестная ошибка'}`)
        }

        const data = await response.json()
        const foundProducts = data.products || []
        
        if (foundProducts.length > 0) {
          totalProducts = [...totalProducts, ...foundProducts]
        }

        setFileStatuses(prev => prev.map((fs, idx) => 
          idx === i ? { ...fs, status: 'done', productsFound: foundProducts.length } : fs
        ))

        setProcessedFiles(i + 1)
        setProgress(Math.round(((i + 1) / files.length) * 100))
        setStatusMessage(`Обработано ${i + 1} из ${files.length} файлов, найдено ${formatNumber(totalProducts.length)} товаров`)

        if (i < files.length - 1) {
          setStatusMessage(`Пауза перед следующим файлом...`)
          await new Promise(resolve => setTimeout(resolve, 1000))
        }

      } catch (err: any) {
        setFileStatuses(prev => prev.map((fs, idx) => 
          idx === i ? { ...fs, status: 'error', productsFound: 0 } : fs
        ))
        
        if (err.name === 'AbortError') {
          setError(`Превышено время ожидания для файла ${file.name}`)
        } else {
          setError(err instanceof Error ? err.message : 'Неизвестная ошибка')
        }
        setIsLoading(false)
        return
      }
    }

    const endTime = Date.now()
    setProcessingTime((endTime - startTime) / 1000)
    setProducts(totalProducts)
    setStatusMessage(`Готово! Найдено ${formatNumber(totalProducts.length)} товаров`)
    setCurrentFile('')
    setIsLoading(false)
    setProgress(100)
  }

  const handleExportCSV = () => {
    if (products.length === 0) return

    const header = 'Товар,Цена,Валюта'
    const rows = products.map((p) => {
      const currency = p.currency || 'руб'
      return `${p.product_name},${p.price},${currency}`
    })
    const csvContent = [header, ...rows].join('\n')

    const BOM = '\uFEFF'
    const blob = new Blob([BOM + csvContent], { type: 'text/csv;charset=utf-8;' })

    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = `parsed_products_${new Date().toISOString().slice(0,10)}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(link.href)
  }

  const getStatusBadge = (status: string) => {
    switch(status) {
      case 'pending':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-white/5 text-gray-400 text-xs font-medium">
            <Clock className="w-3 h-3" />
            В очереди
          </span>
        )
      case 'processing':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-purple-500/20 text-purple-400 text-xs font-medium">
            <Loader2 className="w-3 h-3 animate-spin" />
            Обработка
          </span>
        )
      case 'done':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-green-500/20 text-green-400 text-xs font-medium">
            <CheckCircle className="w-3 h-3" />
            Готово
          </span>
        )
      case 'error':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-red-500/20 text-red-400 text-xs font-medium">
            <AlertCircle className="w-3 h-3" />
            Ошибка
          </span>
        )
      default:
        return null
    }
  }

  const getFileIcon = (name: string) => {
    const ext = name.split('.').pop()?.toLowerCase()
    if (ext === 'pdf') return <File className="w-5 h-5 text-red-400" />
    if (ext === 'csv') return <FileSpreadsheet className="w-5 h-5 text-green-400" />
    if (ext === 'txt') return <File className="w-5 h-5 text-gray-400" />
    return <FileSpreadsheet className="w-5 h-5 text-purple-400" />
  }

  return (
    <div className="min-h-screen bg-gradient-page" style={{ padding: '20px' }}>
      {/* 🔥 КОНТЕЙНЕР С ПРИНУДИТЕЛЬНЫМИ ОТСТУПАМИ */}
      <div style={{ maxWidth: '1280px', margin: '0 auto', padding: '20px 16px' }}>
        
        {/* Header */}
        <header className="text-center mb-12">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-[#764C94] to-[#a78bfa] shadow-lg shadow-[#764C94]/30 mb-6">
            <Sparkles className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-4xl md:text-5xl font-bold tracking-tight text-white mb-4">
            AI <span className="bg-gradient-to-r from-[#764C94] to-[#a78bfa] bg-clip-text text-transparent">Price Parser</span>
          </h1>
          <p className="text-lg text-gray-400 max-w-2xl mx-auto">
            Загрузите прайс-листы в Excel, CSV или PDF — искусственный интеллект 
            автоматически извлечёт товары, цены и структуру данных
          </p>
        </header>

        {/* Main Card */}
        <div className="bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-6 md:p-8">
          {/* Drop Zone */}
          <div
            {...getRootProps()}
            className={`
              border-2 border-dashed rounded-xl p-8 md:p-12 text-center cursor-pointer transition-all
              ${isDragActive ? 'border-purple-400 bg-purple-500/10 scale-[0.99]' : 'border-white/20 hover:border-purple-400/50 hover:bg-white/5'}
              ${isLoading ? 'opacity-50 pointer-events-none' : ''}
            `}
          >
            <input {...getInputProps()} />
            <div className="flex flex-col items-center gap-4">
              <div className={`
                w-16 h-16 rounded-2xl flex items-center justify-center
                bg-gradient-to-br from-[#764C94] to-[#a78bfa]
                shadow-lg shadow-[#764C94]/30
                transition-transform duration-300
                ${isDragActive ? 'scale-110' : 'group-hover:scale-105'}
              `}>
                <Upload className="w-7 h-7 text-white" />
              </div>
              <div className="text-center">
                <p className="text-lg font-semibold text-white mb-1">
                  {isDragActive ? 'Отпустите файлы для загрузки' : 'Перетащите файлы сюда'}
                </p>
                <p className="text-sm text-gray-400">
                  или кликните для выбора файлов
                </p>
              </div>
              <div className="flex flex-wrap gap-2 justify-center">
                {['.xlsx', '.xls', '.xlsm', '.csv', '.pdf', '.txt'].map(ext => (
                  <span 
                    key={ext}
                    className="px-3 py-1 bg-white/5 text-gray-400 text-xs font-medium rounded-full"
                  >
                    {ext}
                  </span>
                ))}
              </div>
            </div>
          </div>

          {/* File List */}
          {files.length > 0 && (
            <div className="mt-6 space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-400">
                  Файлы: {files.length}
                </span>
                <button
                  onClick={clearAll}
                  className="text-sm text-gray-500 hover:text-red-400 transition-colors font-medium"
                  disabled={isLoading}
                >
                  Очистить все
                </button>
              </div>
              
              <div className="space-y-2">
                {files.map((file, index) => (
                  <div
                    key={index}
                    className="flex items-center justify-between p-4 bg-white/5 rounded-xl border border-white/5 hover:border-white/10 transition-colors"
                  >
                    <div className="flex items-center gap-4 min-w-0 flex-1">
                      <div className="flex-shrink-0">
                        {getFileIcon(file.name)}
                      </div>
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-medium text-white truncate">
                          {file.name}
                        </p>
                        <div className="flex items-center gap-3 mt-1">
                          <span className="text-xs text-gray-500">
                            {formatFileSize(file.size)}
                          </span>
                          {fileStatuses[index]?.status === 'done' && (
                            <span className="text-xs font-medium text-green-400">
                              {formatNumber(fileStatuses[index].productsFound)} товаров
                            </span>
                          )}
                        </div>
                      </div>
                      <div className="flex-shrink-0">
                        {getStatusBadge(fileStatuses[index]?.status || 'pending')}
                      </div>
                    </div>
                    <button
                      onClick={() => removeFile(index)}
                      className="ml-4 p-2 text-gray-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all"
                      disabled={isLoading}
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Action Buttons */}
          {files.length > 0 && !isLoading && (
            <div className="mt-6 flex flex-col sm:flex-row gap-3">
              <button
                onClick={handleUpload}
                className="flex-1 inline-flex items-center justify-center gap-3 px-6 py-3 bg-gradient-to-r from-[#764C94] to-[#a78bfa] text-white font-medium rounded-xl shadow-lg shadow-[#764C94]/30 hover:shadow-[#764C94]/50 hover:scale-[1.02] active:scale-[0.98] transition-all duration-200"
              >
                <Upload className="w-5 h-5" />
                Обработать {files.length} файлов
              </button>
              <button
                onClick={clearAll}
                className="inline-flex items-center justify-center gap-2 px-6 py-3 bg-white/5 text-gray-300 font-medium rounded-xl hover:bg-white/10 transition-all duration-200"
              >
                <Trash2 className="w-4 h-4" />
                Очистить
              </button>
            </div>
          )}

          {/* Progress Section */}
          {isLoading && (
            <div className="mt-6 space-y-4">
              <div className="space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <span className="font-medium text-gray-300">{statusMessage}</span>
                  <span className="font-bold text-purple-400">{progress}%</span>
                </div>
                <div className="h-3 bg-white/5 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-[#764C94] to-[#a78bfa] rounded-full transition-all duration-500 ease-out"
                    style={{ width: `${progress}%` }}
                  />
                </div>
              </div>

              {currentFile && (
                <div className="flex items-center gap-4 p-4 bg-purple-500/10 rounded-xl border border-purple-500/20">
                  <FileSpreadsheet className="w-5 h-5 text-purple-400 flex-shrink-0" />
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-purple-300 truncate">
                      {currentFile}
                    </p>
                    <div className="flex items-center gap-4 mt-1">
                      <span className="flex items-center gap-1.5 text-xs text-purple-400">
                        <Loader2 className="w-3 h-3 animate-spin" />
                        Обработка
                      </span>
                      <span className="text-xs text-purple-400">
                        Файл {processedFiles + 1} из {totalFiles}
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center gap-1.5 px-2.5 py-1 bg-white/5 rounded-lg border border-purple-500/20">
                    <Brain className="w-4 h-4 text-purple-400" />
                    <span className="text-xs font-medium text-purple-400">AI</span>
                  </div>
                </div>
              )}

              <div className="flex flex-wrap gap-4 pt-2">
                <div className="flex items-center gap-2 text-sm text-gray-400">
                  <Database className="w-4 h-4 text-gray-500" />
                  <span>Товаров: <span className="font-semibold text-white">{formatNumber(products.length)}</span></span>
                </div>
                <div className="flex items-center gap-2 text-sm text-gray-400">
                  <Layers className="w-4 h-4 text-gray-500" />
                  <span>Файлов: <span className="font-semibold text-white">{processedFiles}/{totalFiles}</span></span>
                </div>
              </div>
            </div>
          )}

          {/* Error Message */}
          {error && (
            <div className="mt-4 p-4 bg-red-500/10 border border-red-500/20 rounded-xl flex items-start gap-3">
              <AlertCircle className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
              <div>
                <p className="font-medium text-red-400">Ошибка</p>
                <p className="text-sm text-red-400/80 mt-0.5">{error}</p>
              </div>
            </div>
          )}

          {/* Success Message */}
          {processingTime > 0 && products.length > 0 && !isLoading && (
            <div className="mt-4 p-4 bg-green-500/10 border border-green-500/20 rounded-xl flex items-center gap-3">
              <CheckCircle className="w-5 h-5 text-green-400 flex-shrink-0" />
              <div>
                <p className="font-medium text-green-400">Готово!</p>
                <p className="text-sm text-green-400/80 mt-0.5">
                  Найдено {formatNumber(products.length)} товаров за {processingTime.toFixed(1)} сек
                </p>
              </div>
            </div>
          )}
        </div>

        {/* Results Table */}
        {products.length > 0 && (
          <div className="mt-8 bg-white/5 backdrop-blur-xl border border-white/10 rounded-2xl p-6 md:p-8">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
              <div>
                <h2 className="text-xl font-semibold text-white">
                  Результаты парсинга
                </h2>
                <p className="text-sm text-gray-400 mt-1">
                  Найдено {formatNumber(products.length)} товаров из {files.length} файлов
                </p>
              </div>
              <button
                onClick={handleExportCSV}
                className="inline-flex items-center justify-center gap-2 px-5 py-2.5 bg-[#764C94] text-white font-medium rounded-xl hover:bg-[#8b6ba8] transition-colors shadow-lg shadow-[#764C94]/30"
              >
                <Download className="w-4 h-4" />
                Скачать CSV
              </button>
            </div>

            <div className="overflow-x-auto rounded-xl border border-white/5">
              <table className="w-full">
                <thead className="bg-white/5">
                  <tr>
                    <th className="text-left py-4 px-5 text-xs font-semibold text-gray-400 uppercase tracking-wider">
                      №
                    </th>
                    <th className="text-left py-4 px-5 text-xs font-semibold text-gray-400 uppercase tracking-wider">
                      Товар / Услуга
                    </th>
                    <th className="text-right py-4 px-5 text-xs font-semibold text-gray-400 uppercase tracking-wider">
                      Цена
                    </th>
                    <th className="text-left py-4 px-5 text-xs font-semibold text-gray-400 uppercase tracking-wider">
                      Валюта
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5">
                  {products.slice(0, 50).map((product, index) => (
                    <tr key={index} className="hover:bg-white/5 transition-colors">
                      <td className="py-4 px-5 text-sm text-gray-500">
                        {index + 1}
                      </td>
                      <td className="py-4 px-5 text-sm font-medium text-white">
                        {product.product_name}
                      </td>
                      <td className="py-4 px-5 text-sm text-right font-semibold text-white">
                        {product.price.toFixed(2)}
                      </td>
                      <td className="py-4 px-5 text-sm text-gray-400">
                        {product.currency || 'руб'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {products.length > 50 && (
                <div className="p-4 text-center text-sm text-gray-500 bg-white/5 border-t border-white/5">
                  Показано первые 50 из {formatNumber(products.length)} товаров
                </div>
              )}
            </div>
          </div>
        )}

        {/* Footer */}
        <footer className="mt-8 text-center text-sm text-gray-500">
          <p>
            AI Price Parser • Интеллектуальная обработка прайс-листов
          </p>
        </footer>
        
      </div>
    </div>
  )
}