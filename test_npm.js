const fs = require('fs');
const json = {
    "sticker-pack-id": "com.snowcorp.stickerly.android.stickercontentprovider b5e7275f-f1de-4137-961f-57becfad34f2",
    "sticker-pack-name": "My Pack",
    "sticker-pack-publisher": "My Author",
    "emojis": ["🤖"]
};
const exifAttr = Buffer.from([0x49, 0x49, 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00, 0x01, 0x00, 0x41, 0x57, 0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x16, 0x00, 0x00, 0x00]);
const jsonBuff = Buffer.from(JSON.stringify(json), "utf-8");
const exif = Buffer.concat([exifAttr, jsonBuff]);
exif.writeUIntLE(jsonBuff.length, 14, 4);

console.log(exif.toString('hex'));
