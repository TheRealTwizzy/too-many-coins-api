const fs = require('fs');
const path = require('path');

const filePath = path.join('c:', 'dev', 'too_many_coins', 'public', 'index.html');
const content = fs.readFileSync(filePath, 'utf-8');
const lines = content.split('\n');

// Find the Legacy Emergency Controls card and fix its indentation
let found = false;
for (let i = 0; i < lines.length; i++) {
  if (lines[i].includes('Legacy Emergency Controls') && !found) {
    console.log(`Found Legacy Emergency Controls at line ${i}`);
    found = true;
    
    // Fix the indentation of this card and everything inside it
    // The card should be at the same indentation level as the Data Reset Operations card
    // which is 5 tabs (for <div class="card">)
    // But it currently has 6 tabs
    
    // We need to go back to find where the legacy card starts
    for (let j = i - 1; j >= Math.max(0, i - 5); j--) {
      if (lines[j].includes('<div class="card">') && lines[j].includes('\t\t\t\t\t\t\t\t\t\t')) {
        console.log(`Found card opening at line ${j}: current indent is 10 tabs`);
        
        // Fix this line and follow lines until </div>
        let cardLevel = 0;
        for (let k = j; k < Math.min(j + 30, lines.length); k++) {
          // Remove one tab from the beginning if it has extra indentation
          if (lines[k].startsWith('\t\t\t\t\t\t\t\t\t\t\t')) {
            // 11 tabs - reduce to 10
            lines[k] = lines[k].substring(1);
            console.log(`Fixed indentation at line ${k}`);
          } else if (lines[k].startsWith('\t\t\t\t\t\t\t\t\t\t')) {
            // 10 tabs - this might be okay for content inside
            // Check if we're at </div> to know when to stop
            if (lines[k].trim() === '</div>' && k > j) {
              // Could be the closing div of the card
              cardLevel++;
            }
          }
          
          if (lines[k].includes('</section>') && cardLevel >= 1) {
            break;
          }
        }
        break;
      }
    }
    break;
  }
}

if (!found) {
  console.log('Legacy card not found');
} else {
  fs.writeFileSync(filePath, lines.join('\n'), 'utf-8');
  console.log('SUCCESS: Indentation fixed');
}
