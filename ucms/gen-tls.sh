openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout key.pem -out cert.pem -subj "/C=US/ST=Utah/L=SLC/O=Example Corp/OU=Testing/CN=g.uvoo.io"
