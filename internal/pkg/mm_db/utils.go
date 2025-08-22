package mm_db

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"gorm.io/gorm"
)

/*
RelevanceField represents the name of the field generated in full-text search queries into databases.
*/
const RelevanceField = "relevance"

/*
GenerateFuzzySearch generates a piece of query for a full-text search on specfic DB fields by indicating
the search key and the relevance threshold to consider. The list of fields takes into account the order
of the fields themselves to give greater weight in similarity ranking.
*/
func GenerateFuzzySearch(query *gorm.DB, searchKey string, fields []string, relevanceThreshold float64) {
	var regexFields []string
	for _, field := range fields {
		regexFields = append(regexFields, fmt.Sprintf("regexp_replace(%s, '[^\\w]+',' ', 'g')", field))
	}
	allFields := append(fields, regexFields...)
	slices.Sort(allFields)
	allFields = slices.Compact(allFields)
	allFieldsText := strings.Join(allFields, " || ' ' || ")

	regx := regexp.MustCompile(`[^\w]+`)
	searchKeyFields := regx.ReplaceAllString(strings.TrimSpace(searchKey), " & ")

	query.Joins(fmt.Sprintf(", to_tsvector('simple', %s) full_text", allFieldsText))
	query.Joins(fmt.Sprintf(", to_tsquery('simple', '%s') query_key", searchKeyFields))
	for i, field := range fields {
		query.Joins(fmt.Sprintf(", NULLIF(ts_rank(to_tsvector(regexp_replace(%s, '[^\\w]+',' ', 'g') || ' ' || %s), query_key), 0) rank_%d", field, field, i))
	}
	query.Joins(fmt.Sprintf(", SIMILARITY('%s', %s) %s", searchKeyFields, allFieldsText, RelevanceField))
	query.Where(fmt.Sprintf("query_key @@ full_text OR %s >= %.2f", RelevanceField, relevanceThreshold))
}

/*
GenerateFuzzySearchOrderQuery generates a piece of query to allow sorting results
based on the ranking on each fields provided by the query during a full-text search.
*/
func GenerateFuzzySearchOrderQuery(fields []string, orderDir OrderDir) string {
	var orderByFields []string
	for i := range fields {
		orderByFields = append(orderByFields, fmt.Sprintf("rank_%d", i))
	}
	orderByFields = append(orderByFields, RelevanceField)
	orderBy := strings.Join(orderByFields, ", ")
	return fmt.Sprintf("%s %s", orderBy, orderDir)
}
