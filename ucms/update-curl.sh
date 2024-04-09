update(){
curl -k -X PATCH https://localhost:18443/page/978e6ab6-8610-49aa-8130-155629f59986 \
  -u username:password \
  -H "Content-Type: application/json" \
  -d '{"title": "patch 2"}'
}
  #$ -d '{"id": "159bcfce-6301-4727-a0bd-99b4446205ce", "title": "patch 1"}'

  #  -d '{"title": "patch 1", "content": "# Test", "template": "markdown"}'
update

