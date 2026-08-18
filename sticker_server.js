const http = require('http');
const ffmpeg = require('fluent-ffmpeg');
ffmpeg.setFfmpegPath(require('ffmpeg-static'));
const fs = require('fs');
const os = require('os');
const path = require('path');

const server = http.createServer(async (req, res) => {
    if (req.method === 'POST' && (req.url === '/sticker' || req.url === '/tray')) {
        let pack = decodeURIComponent(req.headers['x-pack'] || 'B O T');
        let author = decodeURIComponent(req.headers['x-author'] || 'Z E R O');
        
        let body = [];
        req.on('data', chunk => body.push(chunk));
        req.on('end', async () => {
            const buffer = Buffer.concat(body);
            try {
                if (req.url === '/tray') {
                    const sharp = require('sharp');
                    const trayBuffer = await sharp(buffer)
                        .resize(96, 96, { fit: 'contain', background: { r: 0, g: 0, b: 0, alpha: 0 } })
                        .png()
                        .toBuffer();
                    res.writeHead(200, { 'Content-Type': 'image/png' });
                    res.end(trayBuffer);
                    return;
                }

                // Determine if it's a video by checking magic bytes (mp4 usually starts with ftyp, m3u8, etc)
                // For simplicity, we can just try Sticker first, and if it throws pixel limit, we fallback to ffmpeg
                let isMp4 = buffer.length > 8 && buffer.slice(4, 8).toString('ascii') === 'ftyp';
                let isGif = buffer.length > 3 && buffer.slice(0, 3).toString('ascii') === 'GIF';
                let isVideoHeader = req.headers['x-is-video'] === 'true';
                
                let stickerBuffer = buffer;
                if (isVideoHeader || isMp4 || isGif) {
                    const ext = isGif ? '.gif' : '.mp4';
                    const inputPath = path.join(os.tmpdir(), `input_${Date.now()}${ext}`);
                    const outputPath = path.join(os.tmpdir(), `output_${Date.now()}.webp`);
                    fs.writeFileSync(inputPath, buffer);
                    
                    await new Promise((resolve, reject) => {
                        ffmpeg(inputPath)
                            .outputOptions([
                                '-vcodec', 'libwebp',
                                '-vf', "scale=512:512:force_original_aspect_ratio=decrease,format=rgba,pad=512:512:-1:-1:color=#00000000",
                                '-r', '10', // Stable fps
                                '-lossless', '0',
                                '-compression_level', '6', // Max compression
                                '-qscale', '10', // Low quality for safety
                                '-loop', '0',
                                '-preset', 'picture',
                                '-threads', '4',
                                '-an',
                                '-t', '00:00:05.000' // max 5.0 seconds
                            ])
                            .save(outputPath)
                            .on('end', resolve)
                            .on('error', reject);
                    });
                    
                    stickerBuffer = fs.readFileSync(outputPath);
                    fs.unlinkSync(inputPath);
                    fs.unlinkSync(outputPath);
                    
                    // Return pure FFMPEG WebP without EXIF
                    res.writeHead(200, { 'Content-Type': 'image/webp' });
                    res.end(stickerBuffer);
                } else {
                    const sharp = require('sharp');
                    const webpBuffer = await sharp(stickerBuffer)
                        .resize(512, 512, { fit: 'contain', background: { r: 0, g: 0, b: 0, alpha: 0 } })
                        .webp({ quality: 50 })
                        .toBuffer();
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
