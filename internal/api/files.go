package api

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// CompressFilesHandler 压缩文件/文件夹
func CompressFilesHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Files      []string `json:"files"`
		OutputName string   `json:"output_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Files) == 0 {
		http.Error(w, "未选择文件", http.StatusBadRequest)
		return
	}

	// 确定输出文件名
	if req.OutputName == "" {
		req.OutputName = "archive.zip"
	}
	if !strings.HasSuffix(req.OutputName, ".zip") {
		req.OutputName += ".zip"
	}

	// 确定输出路径（放在第一个文件的目录）
	firstFile := req.Files[0]
	outputPath := filepath.Join(filepath.Dir(firstFile), req.OutputName)

	// 创建 ZIP 文件
	zipFile, err := os.Create(outputPath)
	if err != nil {
		http.Error(w, "创建压缩文件失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 添加文件到 ZIP
	for _, filePath := range req.Files {
		if err := addToZip(zipWriter, filePath, filepath.Dir(firstFile)); err != nil {
			http.Error(w, "压缩失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "压缩完成",
		"file":    outputPath,
	})
}

// addToZip 添加文件或目录到 ZIP
func addToZip(zipWriter *zip.Writer, source string, baseDir string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return addDirToZip(zipWriter, source, baseDir)
	}
	return addFileToZip(zipWriter, source, baseDir)
}

// addFileToZip 添加单个文件到 ZIP
func addFileToZip(zipWriter *zip.Writer, filePath string, baseDir string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 计算相对路径
	relPath, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		relPath = filepath.Base(filePath)
	}

	// 创建 ZIP 文件条目
	zipFile, err := zipWriter.Create(relPath)
	if err != nil {
		return err
	}

	// 复制文件内容
	_, err = io.Copy(zipFile, file)
	return err
}

// addDirToZip 添加目录到 ZIP
func addDirToZip(zipWriter *zip.Writer, dirPath string, baseDir string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录本身（只添加文件）
		if info.IsDir() {
			return nil
		}

		return addFileToZip(zipWriter, path, baseDir)
	})
}

// DecompressFileHandler 解压缩文件
func DecompressFileHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		File   string `json:"file"`
		Target string `json:"target"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.File == "" {
		http.Error(w, "未指定文件", http.StatusBadRequest)
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(req.File); err != nil {
		http.Error(w, "文件不存在", http.StatusNotFound)
		return
	}

	// 确定解压目标目录
	if req.Target == "" {
		// 默认解压到同目录下的同名文件夹
		req.Target = strings.TrimSuffix(req.File, filepath.Ext(req.File))
	}

	// 创建目标目录
	if err := os.MkdirAll(req.Target, 0755); err != nil {
		http.Error(w, "创建目标目录失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 解压 ZIP 文件
	if err := unzipFile(req.File, req.Target); err != nil {
		http.Error(w, "解压失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "解压完成",
		"target":  req.Target,
	})
}

// unzipFile 解压 ZIP 文件
func unzipFile(zipPath string, targetDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join(targetDir, file.Name)

		// 检查路径遍历攻击
		if !strings.HasPrefix(path, filepath.Clean(targetDir)+string(os.PathSeparator)) {
			return fmt.Errorf("非法文件路径: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		// 创建文件的父目录
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		// 创建文件
		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// ReadFileHandler 读取文件内容（用于编辑器）
func ReadFileHandler(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "未指定文件", http.StatusBadRequest)
		return
	}

	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		http.Error(w, "文件不存在", http.StatusNotFound)
		return
	}

	if info.IsDir() {
		http.Error(w, "不能读取目录", http.StatusBadRequest)
		return
	}

	// 限制文件大小（10MB）
	if info.Size() > 10*1024*1024 {
		http.Error(w, "文件太大（超过10MB），请使用下载功能", http.StatusBadRequest)
		return
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "读取文件失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"content": string(content),
		"path":    filePath,
		"name":    filepath.Base(filePath),
		"size":    info.Size(),
	})
}

// SaveFileHandler 保存文件内容（用于编辑器）
func SaveFileHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Path == "" {
		http.Error(w, "未指定文件", http.StatusBadRequest)
		return
	}

	// 创建备份
	if _, err := os.Stat(req.Path); err == nil {
		backupPath := req.Path + ".backup"
		if err := copyFile(req.Path, backupPath); err != nil {
			// 备份失败只记录，不中断保存
			fmt.Printf("创建备份失败: %v\n", err)
		}
	}

	// 保存文件
	if err := os.WriteFile(req.Path, []byte(req.Content), 0644); err != nil {
		http.Error(w, "保存文件失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "文件已保存",
	})
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// RenameFileHandler 重命名文件/文件夹
func RenameFileHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OldPath string `json:"old_path"`
		NewName string `json:"new_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.OldPath == "" || req.NewName == "" {
		http.Error(w, "参数不完整", http.StatusBadRequest)
		return
	}

	// 检查源文件是否存在
	if _, err := os.Stat(req.OldPath); err != nil {
		http.Error(w, "文件不存在", http.StatusNotFound)
		return
	}

	// 计算新路径
	newPath := filepath.Join(filepath.Dir(req.OldPath), req.NewName)

	// 检查目标是否已存在
	if _, err := os.Stat(newPath); err == nil {
		http.Error(w, "目标文件/文件夹已存在", http.StatusConflict)
		return
	}

	// 重命名
	if err := os.Rename(req.OldPath, newPath); err != nil {
		http.Error(w, "重命名失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "重命名成功",
		"path":    newPath,
	})
}
