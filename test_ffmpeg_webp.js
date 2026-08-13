const ffmpeg = require('fluent-ffmpeg');
ffmpeg.setFfmpegPath(require('ffmpeg-static'));
const fs = require('fs');

async function convertVideoToWebp(inputBuffer) {
    return new Promise((resolve, reject) => {
        const inputPath = '/tmp/input.mp4';
        const outputPath = '/tmp/output.webp';
        fs.writeFileSync(inputPath, inputBuffer);
        
        ffmpeg(inputPath)
            .outputOptions([
                '-vcodec', 'libwebp',
                '-vf', 'scale=\'min(512,iw)\':\'min(512,ih)\':force_original_aspect_ratio=decrease,pad=512:512:(ow-iw)/2:(oh-ih)/2:color=black@0',
                '-lossless', '0',
                '-qscale', '50',
                '-loop', '0',
                '-preset', 'default',
                '-an', '-vsync', '0',
                '-t', '00:00:10' // max 10 seconds
            ])
            .save(outputPath)
            .on('end', () => {
                const webpBuffer = fs.readFileSync(outputPath);
                resolve(webpBuffer);
            })
            .on('error', (err) => {
                reject(err);
            });
    });
}

async function test() {
    try {
        const input = fs.readFileSync('/tmp/test.mp4');
        const webp = await convertVideoToWebp(input);
        
        const { Sticker } = require('wa-sticker-formatter');
        const sticker = new Sticker(webp, { pack: "Test", author: "Test" });
        await sticker.toBuffer();
        console.log("Success! Generated WebP and added EXIF");
    } catch (e) {
        console.error(e);
    }
}
test();
