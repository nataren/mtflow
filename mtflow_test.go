import (
	"encoding/json"
	"io"
	"strings"
)

func main() {
	const jsonStream = `{
    "app": null,
    "attachments": [],
    "content": {
        "message": 3080629,
        "updated_content": "@c\u00e9sar subprogram mtflow"
    },
    "created_at": "2015-03-26T00:06:01.545Z",
    "event": "message-edit",
    "flow": "a38b50b8-8b3b-42bd-a515-88bf7b1f58bd",
    "id": 3080737,
    "persist": false,
    "sent": 1427328361545,
    "tags": [],
    "user": "76198",
    "uuid": null
}`

	dec := json.NewDecoder(strings.NewReader(jsonStream))
	for {
		var m FlowdockMessage
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("%+v", m)
		}
	}
}