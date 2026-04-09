package note

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func parseNoteFile(path string) (Note, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Note{}, err
	}

	content := string(data)
	metadata, tags, body, err := parseFrontmatter(content)
	if err != nil {
		return Note{}, fmt.Errorf("parse %s: %w", filepath.Base(path), err)
	}

	return Note{
		Path:     path,
		Metadata: metadata,
		Tags:     tags,
		Body:     body,
	}, nil
}

func parseFrontmatter(content string) (map[string]string, []string, string, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	metadata := map[string]string{}
	if !strings.HasPrefix(content, "---\n") {
		return metadata, nil, content, nil
	}

	rest := content[len("---\n"):]
	separator := strings.Index(rest, "\n---\n")
	if separator < 0 {
		return nil, nil, "", fmt.Errorf("missing closing frontmatter separator")
	}

	header := rest[:separator]
	body := rest[separator+len("\n---\n"):]
	var tags []string
	for _, line := range strings.Split(header, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			return nil, nil, "", fmt.Errorf("invalid frontmatter line %q", line)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "tags" {
			tags = parseTags(value)
			continue
		}
		metadata[key] = strings.Trim(value, `"`)
	}
	metadata["tags"] = formatTags(tags)
	return metadata, tags, body, nil
}

func parseTags(value string) []string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimPrefix(trimmed, "[")
	trimmed = strings.TrimSuffix(trimmed, "]")
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return []string{}
	}
	parts := strings.Split(trimmed, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.Trim(strings.TrimSpace(part), `"`)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}
