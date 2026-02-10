const fs = require('fs');
const path = require('path');

const filePath = path.join('c:', 'dev', 'too_many_coins', 'public', 'index.html');
let content = fs.readFileSync(filePath, 'utf-8');

// Use regex with whitespace flexibility
const oldPatternRegex = /(\s*)<div class="card">\s*<div class="label">Data Reset Operations<\/div>\s*<div class="muted">"I understand this affects all players\."<\/div>\s*<div class="admin-control-strip" style="margin-top:0\.75rem; display:none;" data-admin-controls>[\s\S]*?<\/div>\s*<\/div>/;

// Check if pattern matches
if (oldPatternRegex.test(content)) {
  console.log('Pattern found, attempting replacement');
  
  // For now, let's just do a simpler targeted replacement
  // Replace the line with the old muted text with the new one
  content = content.replace(
    /<div class="muted">"I understand this affects all players\."<\/div>/,
    '<div class="muted">Risk-first resets. Each requires explicit confirmation.<\/div>'
  );
  
  // Replace display:none; inside the Data Reset Operations card
  // We need to be careful to only replace it in the right place
  // Let's use a more specific pattern
  const lines = content.split('\n');
  let inDataResetCard = false;
  let foundCard = false;
  
  for (let i = 0; i < lines.length; i++) {
    if (lines[i].includes('Data Reset Operations')) {
      inDataResetCard = true;
      foundCard = true;
    }
    
    if (inDataResetCard && lines[i].includes('display:none;')) {
      // Remove display:none; from this line
      lines[i] = lines[i].replace(' display:none;', '');
      inDataResetCard = false; // We found and fixed it,stop looking
      break;
    }
    
    // If we hit the closing div of the card, stop
    if (inDataResetCard && lines[i].includes('</div>') && lines[i].trim() === '</div>') {
      // Check if this is the card closing tag
      if (i > 0 && lines[i-1].includes('</div>')) {
        inDataResetCard = false;
      }
    }
  }
  
  if (!foundCard) {
    console.log('ERROR: Could not find Data Reset Operations card');
    process.exit(1);
  }
  
  content = lines.join('\n');
  
  // Now add the second card before </section>
  const sectionPattern = /<\/div>\s*<\/div>\s*<\/section>\s*<!-- closing of admin-emergency -->/;
  const legacyCard = `</div>
						</div>
						<div class="card">
							<div class="label">Legacy Emergency Controls</div>
							<div class="muted">"I understand this affects all players."</div>
							<div class="admin-control-strip" style="margin-top:0.75rem; display:none;" data-admin-controls>
								<div class="row">
									<button type="button" class="danger" disabled>Pause all purchases</button>
									<div class="muted">Stops all star and boost purchases across seasons.</div>
								</div>
								<div class="row">
									<button type="button" class="danger" disabled>Pause all emission</button>
									<div class="muted">Halts emission for all faucets and drips.</div>
								</div>
								<div class="row">
									<button type="button" class="danger" disabled>Freeze all seasons</button>
									<div class="muted">Freezes every season immediately. Requires double-confirmation.</div>
								</div>
							</div>
						</div>
					</section>`;
  
  // Find the closing section for admin-emergency and add the second card
  const adminEmergencyEnd = content.indexOf('</section>');
  if (adminEmergencyEnd > -1) {
    // Find which section this is by looking back for admin-emergency
    let checkPos = adminEmergencyEnd - 1;
    let foundEmergency = false;
    
    // Look back to find which section we're closing
    while (checkPos > 0) {
      if (content.substring(checkPos - 100, checkPos).includes('admin-emergency')) {
        foundEmergency = true;
        break;
      }
      checkPos -= 100;
    }
    
    if (foundEmergency) {
      // Found the right section, insert the second card before </section>
      const newContent = content.substring(0, adminEmergencyEnd) + 
        `<div class="card">
								<div class="label">Legacy Emergency Controls</div>
								<div class="muted">"I understand this affects all players."</div>
								<div class="admin-control-strip" style="margin-top:0.75rem; display:none;" data-admin-controls>
									<div class="row">
										<button type="button" class="danger" disabled>Pause all purchases</button>
										<div class="muted">Stops all star and boost purchases across seasons.</div>
									</div>
									<div class="row">
										<button type="button" class="danger" disabled>Pause all emission</button>
										<div class="muted">Halts emission for all faucets and drips.</div>
									</div>
									<div class="row">
										<button type="button" class="danger" disabled>Freeze all seasons</button>
										<div class="muted">Freezes every season immediately. Requires double-confirmation.</div>
									</div>
								</div>
							</div>
						` + content.substring(adminEmergencyEnd);
      
      fs.writeFileSync(filePath, newContent, 'utf-8');
      console.log('SUCCESS: File updated');
    }
  }
} else {
  console.log('ERROR: Main pattern not found');
  process.exit(1);
}
