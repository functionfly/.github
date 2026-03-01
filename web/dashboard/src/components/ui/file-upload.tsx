import { useState, useRef, useCallback } from "react";
import { Upload, X, File, Image, FileText, Archive, Loader2 } from "lucide-react";
import { Button } from "./button";
import { Progress } from "./progress";
import { cn } from "@/lib/utils";

interface FileWithPreview {
  id: string;
  file: File;
  preview?: string;
  progress: number;
  status: "pending" | "uploading" | "success" | "error";
  error?: string;
}

interface FileUploadProps {
  accept?: string;
  maxFiles?: number;
  maxSize?: number; // in bytes
  onFilesChange?: (files: File[]) => void;
  disabled?: boolean;
  className?: string;
}

export function FileUpload({
  accept,
  maxFiles = 5,
  maxSize = 10 * 1024 * 1024, // 10MB default
  onFilesChange,
  disabled = false,
  className,
}: FileUploadProps) {
  const [files, setFiles] = useState<FileWithPreview[]>([]);
  const [isDragging, setIsDragging] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const generateId = () => Math.random().toString(36).substring(2, 9);

  const getFileIcon = (type: string) => {
    if (type.startsWith("image/")) return <Image className="w-5 h-5" />;
    if (type.includes("pdf") || type.includes("text")) return <FileText className="w-5 h-5" />;
    if (type.includes("zip") || type.includes("archive")) return <Archive className="w-5 h-5" />;
    return <File className="w-5 h-5" />;
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const processFiles = useCallback(
    (fileList: FileList | File[]) => {
      const newFiles: FileWithPreview[] = [];
      const filesArray = Array.from(fileList);

      const remainingSlots = maxFiles - files.length;
      if (remainingSlots <= 0) return;

      filesArray.slice(0, remainingSlots).forEach((file) => {
        if (file.size > maxSize) {
          newFiles.push({
            id: generateId(),
            file,
            progress: 0,
            status: "error",
            error: `File too large. Max size is ${formatFileSize(maxSize)}`,
          });
          return;
        }

        const fileWithPreview: FileWithPreview = {
          id: generateId(),
          file,
          progress: 0,
          status: "pending",
        };

        // Generate preview for images
        if (file.type.startsWith("image/")) {
          const reader = new FileReader();
          reader.onload = (e) => {
            setFiles((prev) =>
              prev.map((f) =>
                f.id === fileWithPreview.id
                  ? { ...f, preview: e.target?.result as string }
                  : f
              )
            );
          };
          reader.readAsDataURL(file);
        }

        newFiles.push(fileWithPreview);
      });

      setFiles((prev) => {
        const updated = [...prev, ...newFiles];
        onFilesChange?.(updated.map((f) => f.file));
        return updated;
      });
    },
    [files.length, maxFiles, maxSize, onFilesChange]
  );

  const removeFile = (id: string) => {
    setFiles((prev) => {
      const updated = prev.filter((f) => f.id !== id);
      onFilesChange?.(updated.map((f) => f.file));
      return updated;
    });
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    if (!disabled) setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    if (!disabled && e.dataTransfer.files) {
      processFiles(e.dataTransfer.files);
    }
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      processFiles(e.target.files);
    }
  };

  const handleClick = () => {
    if (!disabled) {
      inputRef.current?.click();
    }
  };

  return (
    <div className={cn("space-y-4", className)}>
      <div
        onClick={handleClick}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        className={cn(
          "border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors",
          isDragging
            ? "border-primary bg-primary/5"
            : "border-border hover:border-primary/50 hover:bg-bg-hover",
          disabled && "opacity-50 cursor-not-allowed"
        )}
      >
        <input
          ref={inputRef}
          type="file"
          accept={accept}
          multiple={maxFiles > 1}
          onChange={handleFileSelect}
          disabled={disabled}
          className="hidden"
        />
        <div className="flex flex-col items-center gap-2">
          <div className="p-3 bg-primary/10 rounded-full">
            <Upload className="w-6 h-6 text-primary" />
          </div>
          <div>
            <p className="text-text-primary font-medium">
              {isDragging ? "Drop files here" : "Drag & drop files here"}
            </p>
            <p className="text-sm text-text-muted">
              or click to browse ({maxFiles} files max, {formatFileSize(maxSize)} each)
            </p>
          </div>
          {accept && (
            <p className="text-xs text-text-muted">Accepted: {accept}</p>
          )}
        </div>
      </div>

      {files.length > 0 && (
        <div className="space-y-2">
          {files.map((file) => (
            <div
              key={file.id}
              className="flex items-center gap-3 p-3 rounded-lg bg-bg-secondary border border-border"
            >
              {file.preview ? (
                <img
                  src={file.preview}
                  alt={file.file.name}
                  className="w-10 h-10 rounded object-cover"
                />
              ) : (
                <div className="p-2 bg-bg-tertiary rounded">
                  {getFileIcon(file.file.type)}
                </div>
              )}

              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between mb-1">
                  <p className="text-sm font-medium text-text-primary truncate">
                    {file.file.name}
                  </p>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6"
                    onClick={() => removeFile(file.id)}
                    disabled={disabled}
                  >
                    <X className="w-4 h-4" />
                  </Button>
                </div>
                <div className="flex items-center gap-2">
                  <p className="text-xs text-text-muted">
                    {formatFileSize(file.file.size)}
                  </p>
                  {file.status === "uploading" && (
                    <div className="flex items-center gap-1 text-xs text-primary">
                      <Loader2 className="w-3 h-3 animate-spin" />
                      Uploading...
                    </div>
                  )}
                  {file.status === "error" && (
                    <p className="text-xs text-red-400">{file.error}</p>
                  )}
                  {file.status === "success" && (
                    <p className="text-xs text-green-400">Uploaded</p>
                  )}
                </div>
                {file.status === "uploading" && (
                  <Progress value={file.progress} className="h-1 mt-1" />
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
