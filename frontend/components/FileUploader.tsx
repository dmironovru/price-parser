'use client'

import { useCallback, useState } from 'react'
import { useDropzone } from 'react-dropzone'
import { Upload, File, X, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { Card, CardContent } from '@/components/ui/card'

interface FileUploaderProps {
  onFileUpload: (file: File) => void
  isUploading: boolean
  progress: number
}

export function FileUploader({ onFileUpload, isUploading, progress }: FileUploaderProps) {
  const [file, setFile] = useState<File | null>(null)

  const onDrop = useCallback((acceptedFiles: File[]) => {
    if (acceptedFiles.length > 0) {
      const selectedFile = acceptedFiles[0]
      setFile(selectedFile)
      onFileUpload(selectedFile)
    }
  }, [onFileUpload])

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'text/csv': ['.csv'],
      'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': ['.xlsx'],
      'application/vnd.ms-excel': ['.xls'],
      'application/pdf': ['.pdf'],
      'text/plain': ['.txt'],
    },
    maxFiles: 1,
    disabled: isUploading,
  })

  const removeFile = () => {
    setFile(null)
  }

  return (
    <div className="w-full max-w-2xl mx-auto">
      <Card>
        <CardContent className="p-6">
          <div
            {...getRootProps()}
            className={`
              border-2 border-dashed rounded-lg p-8 text-center cursor-pointer
              transition-colors duration-200
              ${isDragActive ? 'border-primary bg-primary/10' : 'border-muted-foreground/25'}
              ${isUploading ? 'opacity-50 cursor-not-allowed' : 'hover:border-primary/50'}
            `}
          >
            <input {...getInputProps()} />
            {isUploading ? (
              <Loader2 className="w-12 h-12 mx-auto text-primary animate-spin" />
            ) : (
              <Upload className="w-12 h-12 mx-auto text-muted-foreground" />
            )}
            <h3 className="mt-4 text-lg font-semibold">
              {isDragActive ? 'Отпустите файл для загрузки' : 'Перетащите файл сюда'}
            </h3>
            <p className="mt-2 text-sm text-muted-foreground">
              или кликните для выбора файла
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              Поддерживаемые форматы: CSV, Excel, PDF, TXT
            </p>
          </div>

          {file && !isUploading && (
            <div className="mt-4 p-3 bg-muted rounded-lg flex items-center justify-between">
              <div className="flex items-center gap-3">
                <File className="w-5 h-5 text-primary" />
                <div>
                  <p className="text-sm font-medium">{file.name}</p>
                  <p className="text-xs text-muted-foreground">
                    {(file.size / 1024).toFixed(1)} KB
                  </p>
                </div>
              </div>
              <Button variant="ghost" size="sm" onClick={removeFile}>
                <X className="w-4 h-4" />
              </Button>
            </div>
          )}

          {isUploading && (
            <div className="mt-4">
              <Progress value={progress} className="h-2" />
              <p className="mt-2 text-sm text-center text-muted-foreground">
                Обработка... {progress}%
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
