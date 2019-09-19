package simple_search

import (
	"strings"
)

type HTMLExtractor struct {
	ac *ACAutomaton
}

func NewHTMLExtractor() *HTMLExtractor {
	ac := NewACAutomaton([]string{"<option", "<script", "<style", "<meta", "<img"})
	return &HTMLExtractor{ac: ac}
}

func (extractor *HTMLExtractor) Extract(html string) (text string, links []string) {
	// 完全删除部分标签及其包含的内容
	poses := extractor.ac.Find(html)
	deletePos := make([][]int, 0)
	for _, res := range poses {
		endTag := "</" + html[res.pos+1:res.pos+res.length] + ">"
		index := strings.Index(html[res.pos:], endTag)
		var poses []int
		if index == -1 {
			// 若没有结束标签，则查找>
			index = strings.Index(html[res.pos:], ">")
			poses = []int{res.pos, res.pos + index + 1}
		} else {
			poses = []int{res.pos, res.pos + index + len(endTag)}
		}
		deletePos = append(deletePos, poses)
	}
	// 删除deletePos的东西
	deleteCount := 0
	for i := 0; i < len(deletePos); i++ {
		html = html[:deletePos[i][0]-deleteCount] + html[deletePos[i][1]-deleteCount:]
		deleteCount += deletePos[i][1] - deletePos[i][0]
	}

	// 删除标签，保留内容
	for i := 0; i < len(html); {
		if html[i] == '<' {
			// 去除标签
			end := 0
			for end = i + 1; end < len(html) && html[end] != '>'; end++ {
			}
			// 这个是链接标签，可以处理一下
			if html[i+1] == 'a' {
				link := extractLink(html[i:end])
				// 去除无用的链接
				if link != "" && link != "javascript:void(0);" && link != "#" {
					links = append(links, link)
				}
			}

			if end == len(html) {
				html = html[:i]
			} else {
				html = html[:i] + html[end+1:]
			}
		} else {
			i++
		}
	}

	text = html
	return
}

// 从<a>标签中提取/
func extractLink(aTag string) string {
	href := strings.Index(aTag, "href")
	if href == -1 {
		return ""
	}
	end := 0
	for end = href + 6; end < len(aTag); end++ {
		if aTag[end] == '"' {
			break
		}
	}
	return aTag[href+6 : end]
}
