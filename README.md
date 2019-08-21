### Blockchain

```
go run blockchain.go

```

#### Create Blocks

```
$ curl -X POST http://localhost:3000/new \
	-H "Content-Type: application/json" \
	-d '{"title": "Sample Book", "author":"John Doe", "isbn":"909090","publish_date":"2018-05-26"}'

$ curl -X POST http://localhost:3000 \
	-H "Content-Type: application/json" \
	-d '{"book_id": "generated_id", "user": "Mary Doe", "checkout_date":"2018-05-28"}'
```