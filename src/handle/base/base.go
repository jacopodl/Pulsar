package base

func SplitInputBuffer(buf []byte, baseChunk int, paddingRune byte) (int, bool) {
	for i := baseChunk; i < len(buf); i += baseChunk {
		if buf[i-1] == paddingRune {
			return i, true
			break
		}
	}
	return len(buf), len(buf)%baseChunk == 0
}
