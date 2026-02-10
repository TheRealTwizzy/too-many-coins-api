const fs = require('fs');
const path = require('path');

const filePath = path.join('c:', 'dev', 'too_many_coins', 'public', 'index.html');
const content = fs.readFileSync(filePath, 'utf-8');
const lines = content.split('\n');

// Find and fix the legacy card indentation
let found = false;
for (let i = 0; i < lines.length; i++) {
  if (lines[i].includes('Legacy Emergency Controls') && !found) {
    console.log(`Found Legacy Emergency Controls at line ${i}`);
    found = true;
    
    // Go back to find the opening <div class="card">
    for (let j = i - 1; j >= Math.max(i - 2, 0); j--) {
      if (lines[j].includes('<div class="card">')) {
        console.log(`Found card opening at line ${j}`);
        
        // Now fix the indentation from this line forward
        // Count the tabs in the first card below to use as a reference
        // The first card should have the correct indentation
        
        // Go through and remove one tab from lines that have extra indentation
        let bracketDepth = 0;
        for (let k = j; k < Math.min(j + 30, lines.length); k++) {
          const line = lines[k];
          
          // Check if line starts with too many tabs
          if (line.startsWith('\t\t\t\t\t\t\t\t\t\t\t\t')) {
            // 12 tabs, should be 11
            lines[k] = line.substring(1);
          } else if (k === j && line.startsWith('\t\t\t\t\t\t\t\t\t\t')) {
            // The card opening div has 10 tabs, should have 9
            lines[k] = line.substring(1);
          }
          
          if (line.trim() === '</div>') {
            bracketDepth++;
          }
          if (lines[k].includes('</section>')) {
            break;
          }
        }
        break;
      }
    }
    break;
  }
}

if (found) {
  fs.writeFileSync(filePath, lines.join('\n'), 'utf-8');
  console.log('SUCCESS');
} else {
  console.log('FAIL');
}
