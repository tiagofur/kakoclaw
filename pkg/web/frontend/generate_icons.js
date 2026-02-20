import sharp from 'sharp';
import fs from 'fs';

const svg = fs.readFileSync('public/favicon.svg');

async function createIcons() {
  await sharp(svg)
    .resize(192, 192)
    .toFile('public/pwa-192x192.png');
  console.log('Created 192x192');

  await sharp(svg)
    .resize(512, 512)
    .toFile('public/pwa-512x512.png');
  console.log('Created 512x512');
}

createIcons().catch(console.error);
