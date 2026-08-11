const { Sticker, StickerTypes } = require('wa-sticker-formatter');
const http = require('http');

const server = http.createServer(async (req, res) => {
    if (req.method === 'POST' && req.url === '/sticker') {
        let pack = req.headers['x-pack'] || 'B O T';
        let author = req.headers['x-author'] || 'Z E R O';
        
        let body = [];
        req.on('data', chunk => body.push(chunk));
        req.on('end', async () => {
            const buffer = Buffer.concat(body);
            try {
                const sticker = new Sticker(buffer, {
                    pack: pack,
                    author: author,
                    type: StickerTypes.DEFAULT,
                    quality: 50
                });
                const webpBuffer = await sticker.toBuffer();
                res.writeHead(200, { 'Content-Type': 'image/webp' });
                res.end(webpBuffer);
            } catch (err) {
                console.error(err);
                res.writeHead(500);
                res.end(err.toString());
            }
        });
    } else {
        res.writeHead(404);
        res.end();
    }
});

server.listen(4321, '127.0.0.1', () => {
    console.log('Sticker server listening on port 4321');
});
