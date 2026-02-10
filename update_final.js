const fs = require('fs');
const path = require('path');

const filePath = path.join('c:', 'dev', 'too_many_coins', 'public', 'index.html');
const content = fs.readFileSync(filePath, 'utf-8');
const lines = content.split('\n');

// Find the line with Data Reset Operations
let found = false;
for (let i = 0; i < lines.length; i++) {
  if (lines[i].includes('Data Reset Operations')) {
    console.log(`Found Data Reset Operations at line ${i}`);
    found = true;
    
    // Line i is the label
    // Line i+1 should be the muted description
    if (lines[i + 1].includes('"I understand this affects all players."')) {
      console.log('Found old description, replacing...');
      lines[i + 1] = lines[i + 1].replace(
        '"I understand this affects all players."',
        'Risk-first resets. Each requires explicit confirmation.'
      );
    }
    
    // Line i+2 should be the admin-control-strip
    if (lines[i + 2].includes('display:none;')) {
      console.log('Found display:none, removing...');
      lines[i + 2] = lines[i + 2].replace(' display:none;', '');
    }
    
    // Now we need to add status divs after each row
    // Looking at the structure: lines are [row][button][muted][/row]
    // We need to add status divs after [/row]
    
    // Find the first /row after i+3
    let rowCount = 0;
    let needsStatusDiv = [];
    
    for (let j = i + 3; j < Math.min(i + 50, lines.length); j++) {
      if (lines[j].includes('<div class="row"') && !lines[j].includes('</div>')) {
        rowCount++;
      }
      if (lines[j].trim() === '</div>' && rowCount > 0) {
        // Check if this closes a row
        if (j > 0 && lines[j-1].includes('<div class="muted">')) {
          // This closes the row containing a button and description
          if (rowCount === 1) {
            needsStatusDiv.push(j); // After first row, add telemetry status
          } else if (rowCount === 2) {
            needsStatusDiv.push(j); // After second row, add seasonal stats status
          } else if (rowCount === 3) {
            needsStatusDiv.push(j); // After third row, add season status
            break; // We found all three
          }
          rowCount = 0; // Reset for next row
        }
      }
    }
    
    // Insert status divs in reverse order to maintain line numbers
    const indent = '\t\t\t\t\t\t\t\t\t';
    const statusDivs = [
      `${indent}<div id="admin-reset-season-status" class="label" style="margin-top:0.5rem;"></div>`,
      `${indent}<div id="admin-reset-seasonal-stats-status" class="label" style="margin-top:0.5rem;"></div>`,
      `${indent}<div id="admin-reset-telemetry-status" class="label" style="margin-top:0.5rem;"></div>`
    ];
    
    // Wait, I realized I already added the status divs during earlier replacements
    // Let me check if they exist
    let hasStatusDivs = false;
    for (let j = i; j < Math.min(i + 50, lines.length); j++) {
      if (lines[j].includes('admin-reset-telemetry-status')) {
        hasStatusDivs = true;
        console.log('Status divs already exist');
        break;
      }
    }
    
    // Skip the status div insertion since they're already there
    
    // Now add the second card before </section>
    // Find the </section> that closes admin-emergency
    for (let j = i; j < Math.min(i + 50, lines.length); j++) {
      if (lines[j].trim() === '</section>') {
        console.log(`Found closing section at line ${j}`);
        
        const legacyCard = `\t\t\t\t\t\t\t\t<div class="card">
\t\t\t\t\t\t\t\t\t<div class="label">Legacy Emergency Controls</div>
\t\t\t\t\t\t\t\t\t<div class="muted">"I understand this affects all players."</div>
\t\t\t\t\t\t\t\t\t<div class="admin-control-strip" style="margin-top:0.75rem; display:none;" data-admin-controls>
\t\t\t\t\t\t\t\t\t\t<div class="row">
\t\t\t\t\t\t\t\t\t\t\t<button type="button" class="danger" disabled>Pause all purchases</button>
\t\t\t\t\t\t\t\t\t\t\t<div class="muted">Stops all star and boost purchases across seasons.</div>
\t\t\t\t\t\t\t\t\t\t</div>
\t\t\t\t\t\t\t\t\t\t<div class="row">
\t\t\t\t\t\t\t\t\t\t\t<button type="button" class="danger" disabled>Pause all emission</button>
\t\t\t\t\t\t\t\t\t\t\t<div class="muted">Halts emission for all faucets and drips.</div>
\t\t\t\t\t\t\t\t\t\t</div>
\t\t\t\t\t\t\t\t\t\t<div class="row">
\t\t\t\t\t\t\t\t\t\t\t<button type="button" class="danger" disabled>Freeze all seasons</button>
\t\t\t\t\t\t\t\t\t\t\t<div class="muted">Freezes every season immediately. Requires double-confirmation.</div>
\t\t\t\t\t\t\t\t\t\t</div>
\t\t\t\t\t\t\t\t\t</div>
\t\t\t\t\t\t\t\t</div>`;
        
        lines.splice(j, 0, legacyCard);
        console.log('Legacy card inserted');
        break;
      }
    }
    
    break;
  }
}

if (!found) {
  console.log('ERROR: Could not find Data Reset Operations');
  process.exit(1);
}

// Write the file back
fs.writeFileSync(filePath, lines.join('\n'), 'utf-8');
console.log('SUCCESS: File updated');
