const fs = require('fs');
const path = require('path');

const filePath = path.join('c:', 'dev', 'too_many_coins', 'public', 'index.html');
const content = fs.readFileSync(filePath, 'utf-8');
const lines = content.split('\n');

// Find the line with Data Reset Operations
for (let i = 0; i < lines.length; i++) {
  if (lines[i].includes('Data Reset Operations')) {
    console.log(`Found at line ${i}`);
    
    // Line i+1 is the muted description that needs updating
    // It might contain fancy quotes, so let's match more flexibly
    if (lines[i + 1].includes('muted')) {
      console.log(`Line ${i+1} current: ${lines[i+1]}`);
      
      // Replace the entire muted div
      lines[i + 1] = lines[i + 1].substring(0, lines[i + 1].indexOf('<div class="muted"')) +
        '<div class="muted">Risk-first resets. Each requires explicit confirmation.</div>';
      
      console.log(`Line ${i+1} updated: ${lines[i+1]}`);
    }
    
    break;
  }
}

fs.writeFileSync(filePath, lines.join('\n'), 'utf-8');
console.log('SUCCESS');
