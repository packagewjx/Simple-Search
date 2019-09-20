package simple_search

import (
	sort "sort"
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

	// 删除标签，保留内容
	for i := 0; i < len(html); {
		if html[i] == '<' {
			// 去除标签
			end := 0
			for end = i + 1; end < len(html) && html[end] != '>'; end++ {
			}
			deletePos = append(deletePos, []int{i, end + 1})
			i = end + 1
		} else {
			i++
		}
	}

	sort.Slice(deletePos, func(i, j int) bool {
		return deletePos[i][0] < deletePos[j][0]
	})

	builder := strings.Builder{}
	// 删除deletePos的东西
	start := 0
	for i := 0; i < len(deletePos); i++ {
		// 有一部分重复的跳过。因为ac自动机与扫描标签的加入了同样的结果
		if start > deletePos[i][0] {
			continue
		}
		builder.WriteString(html[start:deletePos[i][0]])
		// 这个是链接标签，可以处理一下
		if html[deletePos[i][0]+1] == 'a' {
			link := extractLink(html[deletePos[i][0]:deletePos[i][1]])
			// 去除无用的链接
			if link != "" && link != "javascript:void(0);" && link != "#" {
				links = append(links, link)
			}
		}
		start = deletePos[i][1]
	}
	// 把结尾的写进去
	if start < len(html) {
		builder.WriteString(html[start:])
	}

	text = builder.String()
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
