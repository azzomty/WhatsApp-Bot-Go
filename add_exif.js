const { Sticker, StickerTypes } = require('wa-sticker-formatter');
const fs = require('fs');

async function run() {
    try {
        const [inputFile, outputFile, pack, author] = process.argv.slice(2);
        const buffer = fs.readFileSync(inputFile);
        const sticker = new Sticker(buffer, {
            pack: pack || 'B O T',
            author: author || 'Z E R O',
            type: StickerTypes.DEFAULT,
            quality: 50
        });
        const webpBuffer = await sticker.toBuffer();
        fs.writeFileSync(outputFile, webpBuffer);
    } catch(e) {
        console.error(e);
        process.exit(1);
    }
}
run();
