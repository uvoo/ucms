markdown(){
curl -X POST http://localhost:18080/page \
  -u username:password \
  -H "Content-Type: application/json" \
  -d '{"title": "My Markdown File", "content": "## Heading\n\nThis is some **Markdown** content.", "template": "markdown"}'
}

mdboostrap(){
curl -X POST http://localhost:18080/page \
  -u username:password \
  -H "Content-Type: application/json" \
  -d '{"name": "one", "title": "My mdbootstrap", "content": "<h1>Hello mdboostrap!<h1>Smile<i class=\"fas fa-smile\"></i>", "template": "mdbootstrap"}'
}

y(){
curl -X POST http://localhost:18080/pages \
  -H "Content-Type: application/json" \
  -d '{"name": "test this"}'
}
# y
mdboostrap
exit
 curl localhost:8080/markdown/1

