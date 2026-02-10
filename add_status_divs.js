const fs = require('fs');
const path = require('path');

const filePath = path.join('c:', 'dev', 'too_many_coins', 'public', 'index.html');
const content = fs.readFileSync(filePath, 'utf-8');
const lines = content.split('\n');

let found = false;
for (let i = 0; i < lines.length; i++) {
  if (lines[i].includes('Data Reset Operations') && !found) {
    console.log(`Found Data Reset Operations at line ${i}`);
    found = true;
    
    // Look for the first admin-reset-telemetry button and add its status div
    for (let j = i; j < Math.min(i + 30, lines.length); j++) {
      if (lines[j].includes('id="admin-reset-telemetry"')) {
        console.log(`Found telemetry button at line ${j}`);
        // Find the closing </div> of this row
        for (let k = j + 1; k < j + 5; k++) {
          if (lines[k].trim() === '</div>' && !lines[k].includes('class')) {
            // Found the closing div, insert status div after it
            if (!lines[k + 1].includes('admin-reset-telemetry-status')) {
              const indent = '\t\t\t\t\t\t\t\t\t';
              lines.splice(k + 1, 0, `${indent}<div id="admin-reset-telemetry-status" class="label" style="margin-top:0.5rem;"></div>`);
              console.log(`Inserted telemetry status div at line ${k + 1}`);
            }
            break;
          }
        }
        break;
      }
    }
    
    // Look for the reset-seasonal-stats status div and make sure it's there
    for (let j = i; j < Math.min(i + 40, lines.length); j++) {
      if (lines[j].includes('id="admin-reset-seasonal-stats"')) {
        console.log(`Found seasonal stats button at line ${j}`);
        // Find the closing </div> of this row
        for (let k = j + 1; k < j + 5; k++) {
          if (lines[k].trim() === '</div>' && !lines[k].includes('class')) {
            // Check if status div already exists
            if (lines[k + 1] && lines[k + 1].includes('admin-reset-seasonal-stats-status')) {
              console.log('Seasonal stats status div already exists');
            } else {
              const indent = '\t\t\t\t\t\t\t\t\t';
              lines.splice(k + 1, 0, `${indent}<div id="admin-reset-seasonal-stats-status" class="label" style="margin-top:0.5rem;"></div>`);
              console.log(`Inserted seasonal stats status div at line ${k + 1}`);
            }
            break;
          }
        }
        break;
      }
    }
    
    // Look for the reset-season button and add its status div
    for (let j = i; j < Math.min(i + 50, lines.length); j++) {
      if (lines[j].includes('id="admin-reset-season"')) {
        console.log(`Found season button at line ${j}`);
        // Find the closing </div> of this row
        for (let k = j + 1; k < j + 5; k++) {
          if (lines[k].trim() === '</div>' && !lines[k].includes('class')) {
            // Found the closing div, insert status div after it
            if (!lines[k + 1].includes('admin-reset-season-status')) {
              const indent = '\t\t\t\t\t\t\t\t\t';
              lines.splice(k + 1, 0, `${indent}<div id="admin-reset-season-status" class="label" style="margin-top:0.5rem;"></div>`);
              console.log(`Inserted season status div at line ${k + 1}`);
            }
            break;
          }
        }
        break;
      }
    }
    
    break;
  }
}

fs.writeFileSync(filePath, lines.join('\n'), 'utf-8');
console.log('SUCCESS');
