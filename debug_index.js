const fs = require('fs');
const path = require('path');

const filePath = path.join('c:', 'dev', 'too_many_coins', 'public', 'index.html');
const content = fs.readFileSync(filePath, 'utf-8');

// Find the line with Data Reset Operations
const lines = content.split('\n');
for (let i = 0; i < lines.length; i++) {
  if (lines[i].includes('Data Reset Operations')) {
    console.log(`Found at line ${i}:`);
    // Print the next 20 lines
    for (let j = i; j < Math.min(i + 20, lines.length); j++) {
      console.log(`${j}: ${JSON.stringify(lines[j])}`);
    }
    break;
  }
}

// Also check for display:none
for (let i = 0; i < lines.length; i++) {
  if (lines[i].includes('display:none')) {
    console.log(`\nDisplay:none found at line ${i}: ${JSON.stringify(lines[i])}`);
  }
}
