package search

// Standard English stopwords dropped from both index and queries.
// Spec ref: §8.
var stopwords = map[string]struct{}{}

func init() {
	for _, w := range []string{
		"a", "an", "and", "are", "as", "at", "be", "been", "but", "by",
		"for", "from", "has", "have", "he", "her", "him", "his", "i",
		"if", "in", "is", "it", "its", "me", "my", "no", "not", "of",
		"on", "or", "our", "she", "so", "that", "the", "their", "them",
		"they", "this", "to", "was", "we", "were", "what", "when",
		"where", "which", "who", "will", "with", "would", "you", "your",
	} {
		stopwords[w] = struct{}{}
	}
}
