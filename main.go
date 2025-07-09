package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CodeExtensions defines file extensions to consider as code files
var CodeExtensions = map[string]bool{
	".go":    true,
	".js":    true,
	".ts":    true,
	".jsx":   true,
	".tsx":   true,
	".java":  true,
	".c":     true,
	".cpp":   true,
	".cc":    true,
	".h":     true,
	".hpp":   true,
	".cs":    true,
	".php":   true,
	".rb":    true,
	".py":    true,
	".rs":    true,
	".swift": true,
	".kt":    true,
	".scala": true,
	".sql":   true,
	".html":  true,
	".css":   true,
	".scss":  true,
	".json":  true,
	".yaml":  true,
	".yml":   true,
	".toml":  true,
	".xml":   true,
	".sh":    true,
	".bash":  true,
}

// IgnoreDirs defines directories to skip
var IgnoreDirs = map[string]bool{
	".git":         true,
	".svn":         true,
	"node_modules": true,
	"vendor":       true,
	"build":        true,
	"dist":         true,
	"target":       true,
	"bin":          true,
	"obj":          true,
	".idea":        true,
	".vscode":      true,
	"coverage":     true,
	".next":        true,
	"__pycache__":  true,
}

// FileStats holds statistics for a single file
type FileStats struct {
	TotalLines   int
	CodeLines    int
	BlankLines   int
	CommentLines int
}

// ProjectStats holds statistics for the entire project
type ProjectStats struct {
	FilesByExt map[string]int
	StatsByExt map[string]FileStats
	TotalStats FileStats
	TotalFiles int
}

func main() {
	var projectPath string
	if len(os.Args) > 1 {
		projectPath = os.Args[1]
	} else {
		projectPath = "."
	}

	fmt.Printf("Counting lines of code in: %s\n", projectPath)
	fmt.Println(strings.Repeat("=", 50))

	stats, err := countProjectLines(projectPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printResults(stats)
}

func countProjectLines(rootPath string) (*ProjectStats, error) {
	stats := &ProjectStats{
		FilesByExt: make(map[string]int),
		StatsByExt: make(map[string]FileStats),
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories we want to ignore
		if info.IsDir() {
			if shouldIgnoreDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if it's a code file
		ext := strings.ToLower(filepath.Ext(path))
		if !CodeExtensions[ext] {
			return nil
		}

		// Count lines in the file
		fileStats, err := countLinesInFile(path)
		if err != nil {
			fmt.Printf("Warning: Could not read %s: %v\n", path, err)
			return nil
		}

		// Update statistics
		stats.FilesByExt[ext]++
		stats.TotalFiles++

		extStats := stats.StatsByExt[ext]
		extStats.TotalLines += fileStats.TotalLines
		extStats.CodeLines += fileStats.CodeLines
		extStats.BlankLines += fileStats.BlankLines
		extStats.CommentLines += fileStats.CommentLines
		stats.StatsByExt[ext] = extStats

		stats.TotalStats.TotalLines += fileStats.TotalLines
		stats.TotalStats.CodeLines += fileStats.CodeLines
		stats.TotalStats.BlankLines += fileStats.BlankLines
		stats.TotalStats.CommentLines += fileStats.CommentLines

		return nil
	})

	return stats, err
}

func shouldIgnoreDir(dirName string) bool {
	if IgnoreDirs[dirName] {
		return true
	}
	// Only ignore hidden directories if not "." or ".."
	return dirName != "." && dirName != ".." && strings.HasPrefix(dirName, ".")
}

func countLinesInFile(filePath string) (FileStats, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return FileStats{}, err
	}
	defer file.Close()

	var stats FileStats
	scanner := bufio.NewScanner(file)
	ext := strings.ToLower(filepath.Ext(filePath))

	inBlockComment := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		stats.TotalLines++

		if line == "" {
			stats.BlankLines++
			continue
		}

		// Improved comment detection with block comment support
		switch ext {
		case ".go", ".js", ".ts", ".jsx", ".tsx", ".java", ".c", ".cpp", ".cc", ".h", ".hpp", ".cs", ".php", ".rs", ".swift", ".kt", ".scala", ".css", ".scss", ".sql":
			if inBlockComment {
				stats.CommentLines++
				if strings.Contains(line, "*/") {
					inBlockComment = false
				}
				continue
			}
			if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "--") {
				stats.CommentLines++
				continue
			}
			if strings.HasPrefix(line, "/*") {
				stats.CommentLines++
				if !strings.Contains(line, "*/") {
					inBlockComment = true
				}
				continue
			}
			if strings.HasPrefix(line, "*") {
				stats.CommentLines++
				continue
			}
		case ".py", ".sh", ".bash", ".rb", ".yaml", ".yml", ".toml":
			if strings.HasPrefix(line, "#") {
				stats.CommentLines++
				continue
			}
		case ".html", ".xml":
			if inBlockComment {
				stats.CommentLines++
				if strings.Contains(line, "-->") {
					inBlockComment = false
				}
				continue
			}
			if strings.HasPrefix(line, "<!--") {
				stats.CommentLines++
				if !strings.Contains(line, "-->") {
					inBlockComment = true
				}
				continue
			}
		default:
			// fallback: treat as code
		}

		stats.CodeLines++
	}

	return stats, scanner.Err()
}

func printResults(stats *ProjectStats) {
	// Print summary
	fmt.Printf("Total Files: %d\n", stats.TotalFiles)
	fmt.Printf("Total Lines: %d\n", stats.TotalStats.TotalLines)
	fmt.Printf("Code Lines: %d\n", stats.TotalStats.CodeLines)
	fmt.Printf("Comment Lines: %d\n", stats.TotalStats.CommentLines)
	fmt.Printf("Blank Lines: %d\n", stats.TotalStats.BlankLines)
	fmt.Println()

	// Print breakdown by file extension
	fmt.Println("Breakdown by file type:")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("%-8s %-8s %-10s %-10s %-12s %-10s\n", "Ext", "Files", "Total", "Code", "Comments", "Blank")
	fmt.Println(strings.Repeat("-", 70))

	// Sort extensions for consistent output
	var extensions []string
	for ext := range stats.FilesByExt {
		extensions = append(extensions, ext)
	}
	sort.Strings(extensions)

	for _, ext := range extensions {
		fileCount := stats.FilesByExt[ext]
		extStats := stats.StatsByExt[ext]
		fmt.Printf("%-8s %-8d %-10d %-10d %-12d %-10d\n",
			ext, fileCount, extStats.TotalLines, extStats.CodeLines,
			extStats.CommentLines, extStats.BlankLines)
	}

	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("%-8s %-8d %-10d %-10d %-12d %-10d\n",
		"TOTAL", stats.TotalFiles, stats.TotalStats.TotalLines,
		stats.TotalStats.CodeLines, stats.TotalStats.CommentLines,
		stats.TotalStats.BlankLines)
}
