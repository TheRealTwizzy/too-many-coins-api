const fs = require('fs');
const path = require('path');

const filePath = path.join('c:', 'dev', 'too_many_coins', 'public', 'index.html');
let content = fs.readFileSync(filePath, 'utf-8');

// Define old and new patterns based on CURRENT state
const oldPattern = `							<div class="card">
								<div class="label">Data Reset Operations</div>
								<div class="muted">"I understand this affects all players."</div>
								<div class="admin-control-strip" style="margin-top:0.75rem; display:none;" data-admin-controls>
									<div class="row">
										<button type="button" class="danger" id="admin-reset-telemetry">Reset Telemetry</button>
										<div class="muted">Clears all player telemetry events from the database.</div>
									</div>
									<div class="row" style="margin-top:0.75rem;">
										<button type="button" class="danger" id="admin-reset-seasonal-stats">Reset Seasonal Stats</button>
										<div class="muted">Clears earning and purchase logs for the active season.</div>
									</div>
									<div id="admin-reset-seasonal-stats-status" class="label" style="margin-top:0.5rem;"></div>
									<div class="row" style="margin-top:0.75rem;">
										<button type="button" class="danger" id="admin-reset-season">Reset Season Economy</button>
										<div class="muted">Resets all season economy state (coins, prices, pressure).</div>
									</div>
								</div>
							</div>`;

const newPattern = `							<div class="card">
								<div class="label">Data Reset Operations</div>
								<div class="muted">Risk-first resets. Each requires explicit confirmation.</div>
								<div class="admin-control-strip" style="margin-top:0.75rem;">
									<div class="row">
										<button type="button" class="danger" id="admin-reset-telemetry">Reset Telemetry</button>
										<div class="muted">Clears all player telemetry events from the database.</div>
									</div>
									<div id="admin-reset-telemetry-status" class="label" style="margin-top:0.5rem;"></div>
									<div class="row" style="margin-top:0.75rem;">
										<button type="button" class="danger" id="admin-reset-seasonal-stats">Reset Seasonal Stats</button>
										<div class="muted">Clears earning and purchase logs for the active season.</div>
									</div>
									<div id="admin-reset-seasonal-stats-status" class="label" style="margin-top:0.5rem;"></div>
									<div class="row" style="margin-top:0.75rem;">
										<button type="button" class="danger" id="admin-reset-season">Reset Season Economy</button>
										<div class="muted">Resets all season economy state (coins, prices, pressure).</div>
									</div>
									<div id="admin-reset-season-status" class="label" style="margin-top:0.5rem;"></div>
								</div>
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
							</div>`;

if (content.includes(oldPattern)) {
  content = content.replace(oldPattern, newPattern);
  fs.writeFileSync(filePath, content, 'utf-8');
  console.log('SUCCESS: File updated');
} else {
  console.log('FAILURE: Pattern not found in file');
  
  // Debug: Check if parts of the pattern exist
  if (content.includes('Data Reset Operations')) {
    console.log('DEBUG: Found "Data Reset Operations"');
  }
  if (content.includes('"I understand this affects all players."')) {
    console.log('DEBUG: Found "I understand this affects all players."');
  }
  if (content.includes('display:none;')) {
    console.log('DEBUG: Found "display:none;"');
  }
  
  process.exit(1);
}

