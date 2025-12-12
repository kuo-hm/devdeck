const http = require('http');

const PORT = process.env.PORT || 3001;

const server = http.createServer((req, res) => {
  console.log(`${req.method} ${req.url}`);
  res.statusCode = 200;
  res.setHeader('Content-Type', 'text/plain');
  res.end('Core API Running\n');
});

server.listen(PORT, () => {
  console.log(`Core API listening on port ${PORT}`);
});
