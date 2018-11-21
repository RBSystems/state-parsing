package statequery

import "github.com/byuoitav/state-parser/actions/action"

const (
	//StateQuery is the job name
	StateQuery = "state"
)

/*
QueryJob is a job that will query the state of the static cache and generate events.

This job expects two fields in the input configuration:

QueryString:
The query string, in the format defined below:

S -> SbS | oSc | qaq p qvq

b -> "||" | "&"
o -> (
c -> )
p -> ">" | "=" | "<" | "!="
q -> Literal " character

a -> some field of the queried struct
v -> a literal value


  Basically it provides very basic boolean expressions. AND, OR, and NOT. Nesting is permitted. =, < and > are also permitted for comparison operators. Left hand of comparison operators shoudl always be
  Field names of elements in the record of provided type. Right hand should always be an absolute value.

  For time-based fields you can use a relative time in the format of -15m (fifteen minutes ago). s, m, h, d are acceptable intervals. Each corresponds to Seconds, Minutes, Hours, and Days, respectively.

Cache-Type:
The Cache to query See the cache config for acceptable values.

Data-Type:
The type of record within the cache to query. See the cache config for acceptable values.

*/
type QueryJob struct {
}

//Run .
func (j *QueryJob) Run(context interface{}, actionWrite chan action.Payload) {

}
