const http = require('http');

const PORT = process.env.PORT || 3003;

const server = http.createServer((req, res) => {
  console.log(`${req.method} ${req.url}`);
  res.statusCode = 200;
  res.setHeader('Content-Type', 'text/plain');
  res.end('Worker API Running\n');
});

server.listen(PORT, () => {
  console.log(`Worker API listening on port ${PORT}`);
});
