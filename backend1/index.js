const http = require('http');

const server = http.createServer((req, res) => {
  res.statusCode = 200;
  res.setHeader('Content-Type', 'text/plain');
  res.end('Hello, World!\n');
});

const port = 3001;
server.listen(port, () => {
  for (let i = 0; i < 1000; i++) {
    console.log('Hello, World!');
  }
  console.log(`Server running at http://localhost:${port}/`);
});
