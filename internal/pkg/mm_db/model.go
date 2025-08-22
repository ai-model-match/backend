package mm_db

/*
OrderDir represents the sorting direction can be applied to Database queries.
*/
type OrderDir string

/*
Declaration of sorting directions and assigned value to apply in Database queries.
*/
const (
	Asc  OrderDir = "asc"
	Desc OrderDir = "desc"
)

/*
AvailableOrderDir represents a list of available order directions. It is generally used
inside input DTOs to validate the input parameters provided during an API call.
*/
var AvailableOrderDir = []interface{}{Asc, Desc}
