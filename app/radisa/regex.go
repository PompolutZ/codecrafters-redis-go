package radisa

func SearchKeys(pattern string, keys []string) []string {
    var result []string
    for _, key := range keys {
        if matchesGlob(key, pattern) {
            result = append(result, key)
        }
    }
    return result
}

func matchesGlob(text, pattern string) bool {
    return matchGlobHelper(text, pattern, 0, 0)
}

func matchGlobHelper(text, pattern string, textPos, patternPos int) bool {
    // End of pattern
    if patternPos == len(pattern) {
        return textPos == len(text)
    }
    
    // Handle wildcard
    if pattern[patternPos] == '*' {
        // Try matching 0 characters (skip the *)
        if matchGlobHelper(text, pattern, textPos, patternPos+1) {
            return true
        }
        // Try matching 1 or more characters
        for i := textPos; i < len(text); i++ {
            if matchGlobHelper(text, pattern, i+1, patternPos+1) {
                return true
            }
        }
        return false
    }
    
    // End of text but pattern continues
    if textPos == len(text) {
        return false
    }
    
    // Regular character match
    if pattern[patternPos] == text[textPos] {
        return matchGlobHelper(text, pattern, textPos+1, patternPos+1)
    }
    
    return false
}