package dotsql

import (
	"bufio"
	"regexp"
)

type Scanner struct {
	line    string
	queries map[string]string
	current string
}

type stateFn func(*Scanner) stateFn

func getTag(line string) string {
	re := regexp.MustCompile("^\\s*--\\s*name:\\s*(\\S+)")
	matches := re.FindStringSubmatch(line)
	if matches == nil {
		return ""
	}
	return matches[1]
}

func initialState(s *Scanner) stateFn {
	if tag := getTag(s.line); len(tag) > 0 {
		s.current = tag
		return queryState
	}
	return initialState
}

func queryState(s *Scanner) stateFn {
	if tag := getTag(s.line); len(tag) > 0 {
		s.current = tag
	} else {
		s.appendQueryLine()
	}
	return queryState
}

func (self *Scanner) appendQueryLine() {
	self.queries[self.current] = self.queries[self.current] + self.line + "\n"
}

func (self *Scanner) Run(io *bufio.Scanner) map[string]string {
	self.queries = make(map[string]string)

	for s := initialState; io.Scan(); {
		self.line = io.Text()
		s = s(self)
	}

	return self.queries
}
