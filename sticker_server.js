const { Sticker, StickerTypes } = require('wa-sticker-formatter');
const http = require('http');
const ffmpeg = require('fluent-ffmpeg');
ffmpeg.setFfmpegPath(require('ffmpeg-static'));
const fs = require('fs');
const os = require('os');
const path = require('path');

const server = http.createServer(async (req, res) => {
    if (req.method === 'POST' && req.url === '/sticker') {
        let pack = decodeURIComponent(req.headers['x-pack'] || 'B O T');
        let author = decodeURIComponent(req.headers['x-author'] || 'Z E R O');
        
        let body = [];
        req.on('data', chunk => body.push(chunk));
        req.on('end', async () => {
            const buffer = Buffer.concat(body);
            try {
                // Determine if it's a video by checking magic bytes (mp4 usually starts with ftyp, m3u8, etc)
                // For simplicity, we can just try Sticker first, and if it throws pixel limit, we fallback to ffmpeg
                let isMp4 = buffer.length > 8 && buffer.slice(4, 8).toString('ascii') === 'ftyp';
                let isGif = buffer.length > 3 && buffer.slice(0, 3).toString('ascii') === 'GIF';
                
                let stickerBuffer = buffer;
                if (isMp4 || isGif) {
                    const inputPath = path.join(os.tmpdir(), `input_${Date.now()}${isMp4 ? '.mp4' : '.gif'}`);
                    const outputPath = path.join(os.tmpdir(), `output_${Date.now()}.webp`);
                    fs.writeFileSync(inputPath, buffer);
                    
                    await new Promise((resolve, reject) => {
                        ffmpeg(inputPath)
                            .outputOptions([
                                '-vcodec', 'libwebp',
                                '-vf', "scale='min(512,iw)':'min(512,ih)':force_original_aspect_ratio=decrease,pad=512:512:(ow-iw)/2:(oh-ih)/2:color=black@0",
                                '-r', '10', // 10 fps is fast and enough for stickers
                                '-lossless', '0',
                                '-compression_level', '0', // Fastest encoding speed
                                '-qscale', '30', // Lower quality for faster encoding
                                '-loop', '0',
                                '-preset', 'default',
                                '-threads', '4', // Use multiple threads
                                '-an', '-vsync', '0',
                                '-t', '00:00:07' // max 7 seconds to save time
                            ])
                            .save(outputPath)
                            .on('end', resolve)
                            .on('error', reject);
                    });
                    
                    stickerBuffer = fs.readFileSync(outputPath);
                    fs.unlinkSync(inputPath);
                    fs.unlinkSync(outputPath);
                    
                    // We already encoded to webp, so tell wa-sticker-formatter it's already a webp!
                    // If we pass quality it will re-encode. We just pass the buffer and let it add EXIF.
                    const sticker = new Sticker(stickerBuffer, {
                        pack: pack,
                        author: author,
                        type: StickerTypes.DEFAULT
                    });
                    const webpBuffer = await sticker.toBuffer();
                    res.writeHead(200, { 'Content-Type': 'image/webp' });
                    res.end(webpBuffer);
                } else {
                    const sticker = new Sticker(stickerBuffer, {
                        pack: pack,
                        author: author,
                        type: StickerTypes.DEFAULT,
                        quality: 50
                    });
                    const webpBuffer = await sticker.toBuffer();
                    res.writeHead(200, { 'Content-Type': 'image/webp' });
                    res.end(webpBuffer);
                }
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
