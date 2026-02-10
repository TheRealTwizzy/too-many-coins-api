$filePath = "c:\dev\too_many_coins\public\index.html"
$backupPath = "$filePath.bak"

# Create backup
Copy-Item -Path $filePath -Destination $backupPath -Force

# Read file
$content = [System.IO.File]::ReadAllText($filePath)

$oldPattern = @'
							<div class="card">
								<div class="label">Acknowledgement required</div>
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
'@

$newPattern = @'
							<div class="card">
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
							</div>
'@

if ($content.Contains($oldPattern)) {
    Write-Host "Found pattern, replacing..."
    $newContent = $content.Replace($oldPattern, $newPattern)
    [System.IO.File]::WriteAllText($filePath, $newContent)
    Write-Host "SUCCESS: File updated"
} else {
    Write-Host "FAILURE: Pattern not found in file"
    Exit 1
}
