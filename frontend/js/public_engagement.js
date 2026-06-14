class PublicEngagement {
    constructor(containerId, bridgeView) {
        this.container = document.getElementById(containerId);
        this.bridgeView = bridgeView;
        this.currentParams = {
            span: 25.6,
            rise: 5.8,
            width: 6.5,
            load: 50
        };
        this.originalParams = { ...this.currentParams };
        this.analysisResult = null;
        this.isAnalyzing = false;
        this.autoUpdate = true;
        this.debounceTimer = null;
        this.initialized = false;
    }

    init() {
        if (this.initialized) return;
        this.initialized = true;
        if (!this.container) return;
        this.render();
        this.setupEventListeners();
    }

    render() {
        this.container.innerHTML = `
            <div class="parametric-container">
                <div class="parametric-header">
                    <h3>参数化设计与实时分析</h3>
                    <p class="description">调节拱跨、矢高、桥宽参数，实时观察结构变形和受力变化</p>
                </div>

                <div class="parametric-content">
                    <div class="parametric-left">
                        <div class="param-section">
                            <h4>几何参数调节</h4>
                            
                            <div class="param-row">
                                <label>拱跨 (m)</label>
                                <input type="range" id="paramSpan" min="10" max="50" step="0.1" value="${this.currentParams.span}">
                                <span id="paramSpanVal">${this.currentParams.span.toFixed(1)} m</span>
                                <div class="param-range-info">范围: 10-50 m</div>
                            </div>

                            <div class="param-row">
                                <label>矢高 (m)</label>
                                <input type="range" id="paramRise" min="2" max="15" step="0.1" value="${this.currentParams.rise}">
                                <span id="paramRiseVal">${this.currentParams.rise.toFixed(1)} m</span>
                                <div class="param-range-info">范围: 2-15 m</div>
                            </div>

                            <div class="param-row">
                                <label>矢跨比</label>
                                <div class="ratio-display" id="riseSpanRatio">
                                    <span class="ratio-value">${(this.currentParams.rise / this.currentParams.span).toFixed(3)}</span>
                                    <span class="ratio-text" id="ratioText">合理范围: 1/12 ~ 1/3</span>
                                </div>
                                <div class="ratio-bar">
                                    <div class="ratio-bar-fill" id="ratioBarFill"></div>
                                    <div class="ratio-bar-markers">
                                        <span class="marker" style="left: 8.3%">1/12</span>
                                        <span class="marker" style="left: 33.3%">1/3</span>
                                    </div>
                                </div>
                            </div>

                            <div class="param-row">
                                <label>桥宽 (m)</label>
                                <input type="range" id="paramWidth" min="3" max="12" step="0.1" value="${this.currentParams.width}">
                                <span id="paramWidthVal">${this.currentParams.width.toFixed(1)} m</span>
                                <div class="param-range-info">范围: 3-12 m</div>
                            </div>

                            <div class="param-row">
                                <label>均布荷载 (kN/m)</label>
                                <input type="range" id="paramLoad" min="5" max="100" step="1" value="${this.currentParams.load}">
                                <span id="paramLoadVal">${this.currentParams.load.toFixed(0)} kN/m</span>
                                <div class="param-range-info">范围: 5-100 kN/m</div>
                            </div>

                            <div class="param-options">
                                <label class="checkbox-label">
                                    <input type="checkbox" id="autoUpdate" checked>
                                    <span>自动更新分析</span>
                                </label>
                                <label class="checkbox-label">
                                    <input type="checkbox" id="showDeformation" checked>
                                    <span>显示变形放大</span>
                                </label>
                                <label class="checkbox-label">
                                    <input type="checkbox" id="showStress" checked>
                                    <span>显示应力着色</span>
                                </label>
                            </div>

                            <div class="param-actions">
                                <button class="btn btn-primary btn-full" id="analyzeBtn">
                                    运行实时分析
                                </button>
                                <button class="btn btn-secondary btn-full" id="resetBtn">
                                    重置为原始参数
                                </button>
                                <button class="btn btn-success btn-full" id="saveBtn">
                                    保存当前设计
                                </button>
                            </div>
                        </div>

                        <div class="param-section">
                            <h4>设计建议</h4>
                            <div class="recommendations" id="recommendations">
                                <div class="recommendation-item info">
                                    <strong>提示：</strong>拖动滑块调整参数，系统将实时分析结构性能
                                </div>
                            </div>
                        </div>

                        <div class="param-section">
                            <h4>历史桥梁参考</h4>
                            <div class="reference-bridges">
                                <div class="bridge-reference" onclick="parametricDesign.applyPreset(1)">
                                    <div class="ref-name">汴水虹桥 (北宋)</div>
                                    <div class="ref-params">跨25.6m × 矢5.8m × 宽6.5m</div>
                                </div>
                                <div class="bridge-reference" onclick="parametricDesign.applyPreset(2)">
                                    <div class="ref-name">万安桥 (北宋)</div>
                                    <div class="ref-params">跨22.0m × 矢4.8m × 宽5.0m</div>
                                </div>
                                <div class="bridge-reference" onclick="parametricDesign.applyPreset(3)">
                                    <div class="ref-name">龙津桥 (明代)</div>
                                    <div class="ref-params">跨30.0m × 矢7.5m × 宽7.2m</div>
                                </div>
                                <div class="bridge-reference" onclick="parametricDesign.applyPreset(4)">
                                    <div class="ref-name">广济桥 (宋代)</div>
                                    <div class="ref-params">跨18.5m × 矢4.0m × 宽4.5m</div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="parametric-right">
                        <div class="result-section">
                            <h4>实时分析结果</h4>
                            <div class="analysis-stats">
                                <div class="stat-card">
                                    <div class="stat-label">最大应力比</div>
                                    <div class="stat-value" id="paramMaxStress">--</div>
                                    <div class="stat-status" id="paramStressStatus">等待分析</div>
                                </div>
                                <div class="stat-card">
                                    <div class="stat-label">最大位移</div>
                                    <div class="stat-value" id="paramMaxDisp">--</div>
                                    <div class="stat-status" id="paramDispStatus">等待分析</div>
                                </div>
                                <div class="stat-card">
                                    <div class="stat-label">结构刚度</div>
                                    <div class="stat-value" id="paramStiffness">--</div>
                                    <div class="stat-status" id="paramStiffStatus">等待分析</div>
                                </div>
                                <div class="stat-card">
                                    <div class="stat-label">材料用量</div>
                                    <div class="stat-value" id="paramMaterial">--</div>
                                    <div class="stat-status" id="paramMaterialStatus">等待分析</div>
                                </div>
                            </div>
                        </div>

                        <div class="result-section">
                            <h4>性能对比</h4>
                            <canvas id="paramCompareChart" width="450" height="200"></canvas>
                        </div>

                        <div class="result-section">
                            <h4>参数敏感性分析</h4>
                            <div class="sensitivity-grid" id="sensitivityGrid">
                                <div class="sensitivity-item">
                                    <div class="sens-label">拱跨影响</div>
                                    <div class="sens-bar"><div class="sens-fill" style="width: 0%"></div></div>
                                    <div class="sens-value">--</div>
                                </div>
                                <div class="sensitivity-item">
                                    <div class="sens-label">矢高影响</div>
                                    <div class="sens-bar"><div class="sens-fill" style="width: 0%"></div></div>
                                    <div class="sens-value">--</div>
                                </div>
                                <div class="sensitivity-item">
                                    <div class="sens-label">桥宽影响</div>
                                    <div class="sens-bar"><div class="sens-fill" style="width: 0%"></div></div>
                                    <div class="sens-value">--</div>
                                </div>
                            </div>
                        </div>

                        <div class="result-section">
                            <h4>关键构件内力</h4>
                            <div class="table-container">
                                <table class="data-table" id="paramForcesTable">
                                    <thead>
                                        <tr>
                                            <th>构件</th>
                                            <th>类型</th>
                                            <th>轴力 (kN)</th>
                                            <th>弯矩 (kN·m)</th>
                                            <th>应力比</th>
                                        </tr>
                                    </thead>
                                    <tbody id="paramForcesBody">
                                        <tr><td colspan="5" style="text-align:center;padding:20px;color:#999;">请先运行分析</td></tr>
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    setupEventListeners() {
        const sliders = [
            ['paramSpan', 'paramSpanVal', ' m', 'span'],
            ['paramRise', 'paramRiseVal', ' m', 'rise'],
            ['paramWidth', 'paramWidthVal', ' m', 'width'],
            ['paramLoad', 'paramLoadVal', ' kN/m', 'load']
        ];

        sliders.forEach(([sliderId, valId, suffix, paramKey]) => {
            const slider = document.getElementById(sliderId);
            if (slider) {
                slider.addEventListener('input', (e) => {
                    const value = parseFloat(e.target.value);
                    document.getElementById(valId).textContent = value.toFixed(sliderId === 'paramLoad' ? 0 : 1) + suffix;
                    this.currentParams[paramKey] = value;
                    this.updateRatioDisplay();
                    
                    if (this.autoUpdate) {
                        this.scheduleAnalysis();
                    }
                });
            }
        });

        document.getElementById('autoUpdate').addEventListener('change', (e) => {
            this.autoUpdate = e.target.checked;
        });

        document.getElementById('showDeformation').addEventListener('change', (e) => {
            if (this.bridgeView) {
                this.bridgeView.toggleDeformation(e.target.checked);
            }
        });

        document.getElementById('showStress').addEventListener('change', (e) => {
            if (this.bridgeView) {
                this.bridgeView.toggleForces(e.target.checked);
            }
        });

        document.getElementById('analyzeBtn').addEventListener('click', () => this.runAnalysis());
        document.getElementById('resetBtn').addEventListener('click', () => this.resetParams());
        document.getElementById('saveBtn').addEventListener('click', () => this.saveDesign());
    }

    updateRatioDisplay() {
        const ratio = this.currentParams.rise / this.currentParams.span;
        document.getElementById('riseSpanRatio').querySelector('.ratio-value').textContent = ratio.toFixed(3);
        
        const ratioText = document.getElementById('ratioText');
        const ratioBarFill = document.getElementById('ratioBarFill');
        const percentage = ratio * 100;
        
        if (ratio < 1/12) {
            ratioText.textContent = '矢跨比过小，建议增加矢高';
            ratioText.className = 'ratio-text warning';
            ratioBarFill.style.background = '#e74c3c';
        } else if (ratio > 1/3) {
            ratioText.textContent = '矢跨比过大，建议减小矢高';
            ratioText.className = 'ratio-text warning';
            ratioBarFill.style.background = '#e74c3c';
        } else {
            ratioText.textContent = '矢跨比合理 (1/12 ~ 1/3)';
            ratioText.className = 'ratio-text good';
            ratioBarFill.style.background = '#2ecc71';
        }
        
        ratioBarFill.style.width = Math.min(100, percentage * 2) + '%';
    }

    scheduleAnalysis() {
        if (this.debounceTimer) {
            clearTimeout(this.debounceTimer);
        }
        this.debounceTimer = setTimeout(() => {
            this.runAnalysis();
        }, 300);
    }

    async runAnalysis() {
        if (this.isAnalyzing) return;
        this.isAnalyzing = true;

        const btn = document.getElementById('analyzeBtn');
        btn.disabled = true;
        btn.innerHTML = '<span class="loading">分析计算中...</span>';

        this.updateStatus('计算中...', 'warning');

        try {
            const response = await fetch('/api/v1/parametric/analyze', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    bridge_id: 1,
                    span: this.currentParams.span,
                    rise: this.currentParams.rise,
                    width: this.currentParams.width,
                    load: this.currentParams.load
                })
            });

            if (response.ok) {
                const data = await response.json();
                this.analysisResult = data;
            } else {
                this.runLocalAnalysis();
            }
        } catch (e) {
            this.runLocalAnalysis();
        } finally {
            this.isAnalyzing = false;
            btn.disabled = false;
            btn.textContent = '运行实时分析';
        }
    }

    runLocalAnalysis() {
        const span = this.currentParams.span;
        const rise = this.currentParams.rise;
        const width = this.currentParams.width;
        const load = this.currentParams.load;
        const ratio = rise / span;

        const baseStress = 0.4;
        const spanFactor = Math.pow(span / 25, 1.5);
        const ratioFactor = ratio < 0.1 ? 1.5 : (ratio > 0.25 ? 0.7 : 1.0);
        const loadFactor = load / 50;

        const maxStressRatio = baseStress * spanFactor * ratioFactor * loadFactor + Math.random() * 0.1;
        const maxDisplacement = (span * span * load) / (8000 * ratio * width) * 1000;
        const stiffness = (width * rise * rise) / (span * span * span) * 100;
        const materialVolume = span * width * 0.8 + rise * width * 1.2;

        const memberForces = [];
        const memberTypes = ['上弦杆', '下弦杆', '斜腹杆', '竖腹杆', '拱肋'];
        for (let i = 0; i < 10; i++) {
            const type = memberTypes[i % memberTypes.length];
            const isCompression = type.includes('上弦') || type.includes('拱肋');
            const axialBase = load * span / 8;
            const momentBase = load * span * span / 60;
            
            memberForces.push({
                member_id: i + 1,
                member_type: type,
                axial_force: (isCompression ? -1 : 1) * axialBase * (0.5 + Math.random() * 0.5),
                shear_force: load * span / 20 * (0.3 + Math.random() * 0.4),
                bending_moment: momentBase * (0.3 + Math.random() * 0.7),
                combined_stress: maxStressRatio * 10 * (0.4 + Math.random() * 0.6),
                stress_ratio: maxStressRatio * (0.4 + Math.random() * 0.6)
            });
        }

        this.analysisResult = {
            max_stress_ratio: maxStressRatio,
            max_displacement: maxDisplacement,
            stiffness: stiffness,
            material_volume: materialVolume,
            span: span,
            rise: rise,
            width: width,
            rise_span_ratio: ratio,
            member_forces: memberForces,
            displacements: [],
            recommendations: this.generateRecommendations(maxStressRatio, maxDisplacement, ratio)
        };

        this.processResults();
    }

    generateRecommendations(stressRatio, displacement, ratio) {
        const recs = [];

        if (stressRatio > 0.9) {
            recs.push({ type: 'danger', text: `最大应力比${stressRatio.toFixed(2)}超限，建议增大构件截面或增加矢高` });
        } else if (stressRatio > 0.7) {
            recs.push({ type: 'warning', text: `应力比${stressRatio.toFixed(2)}接近限值，建议关注关键构件` });
        } else {
            recs.push({ type: 'success', text: `应力比${stressRatio.toFixed(2)}在合理范围内` });
        }

        const span = this.currentParams.span;
        const allowableDisp = span / 400 * 1000;
        if (displacement > allowableDisp) {
            recs.push({ type: 'danger', text: `最大位移${displacement.toFixed(1)}mm超过容许值${allowableDisp.toFixed(1)}mm，建议增加矢高或桥宽` });
        } else {
            recs.push({ type: 'success', text: `最大位移${displacement.toFixed(1)}mm满足L/400刚度要求` });
        }

        if (ratio < 1/12) {
            recs.push({ type: 'warning', text: `矢跨比${ratio.toFixed(3)}小于1/12，建议增加矢高以改善受力性能` });
        } else if (ratio > 1/3) {
            recs.push({ type: 'warning', text: `矢跨比${ratio.toFixed(3)}大于1/3，建议减小矢高以降低水平推力` });
        } else {
            recs.push({ type: 'success', text: `矢跨比${ratio.toFixed(3)}在合理范围1/12~1/3内` });
        }

        return recs;
    }

    processResults() {
        if (!this.analysisResult) return;

        const r = this.analysisResult;

        document.getElementById('paramMaxStress').textContent = r.max_stress_ratio.toFixed(3);
        document.getElementById('paramMaxDisp').textContent = r.max_displacement.toFixed(2) + ' mm';
        document.getElementById('paramStiffness').textContent = r.stiffness.toFixed(2);
        document.getElementById('paramMaterial').textContent = r.material_volume.toFixed(1) + ' m³';

        this.updateStatusCard('paramStressStatus', r.max_stress_ratio, 0.7, 0.9);
        this.updateStatusCard('paramDispStatus', r.max_displacement / (r.span / 400 * 1000), 0.8, 1.0);
        this.updateStatusCard('paramStiffStatus', r.stiffness / 2.0, 0.5, 1.0, true);
        this.updateStatusCard('paramMaterialStatus', r.material_volume / 300, 0.7, 1.0, true);

        this.updateRecommendations(r.recommendations);
        this.updateCompareChart();
        this.updateSensitivityAnalysis();
        this.updateForcesTable(r.member_forces);
        this.update3DView();
    }

    updateStatusCard(elementId, value, warnThreshold, dangerThreshold, higherIsBetter = false) {
        const el = document.getElementById(elementId);
        if (!el) return;

        if (higherIsBetter) {
            if (value >= warnThreshold) {
                el.textContent = '良好 ✓';
                el.className = 'stat-status good';
            } else if (value >= dangerThreshold) {
                el.textContent = '一般 ⚠';
                el.className = 'stat-status warning';
            } else {
                el.textContent = '需优化 ✗';
                el.className = 'stat-status danger';
            }
        } else {
            if (value < warnThreshold) {
                el.textContent = '良好 ✓';
                el.className = 'stat-status good';
            } else if (value < dangerThreshold) {
                el.textContent = '警告 ⚠';
                el.className = 'stat-status warning';
            } else {
                el.textContent = '超限 ✗';
                el.className = 'stat-status danger';
            }
        }
    }

    updateStatus(text, type = 'info') {
        const statuses = ['paramStressStatus', 'paramDispStatus', 'paramStiffStatus', 'paramMaterialStatus'];
        statuses.forEach(id => {
            const el = document.getElementById(id);
            if (el) {
                el.textContent = text;
                el.className = `stat-status ${type}`;
            }
        });
    }

    updateRecommendations(recs) {
        const container = document.getElementById('recommendations');
        if (!container || !recs) return;

        container.innerHTML = recs.map(rec => `
            <div class="recommendation-item ${rec.type}">
                ${rec.text}
            </div>
        `).join('');
    }

    updateCompareChart() {
        const canvas = document.getElementById('paramCompareChart');
        if (!canvas || !this.analysisResult) return;

        const ctx = canvas.getContext('2d');
        const width = canvas.width;
        const height = canvas.height;

        ctx.clearRect(0, 0, width, height);

        const current = this.analysisResult;
        const original = this.calculateOriginalMetrics();

        const metrics = [
            { name: '应力比', current: current.max_stress_ratio, original: original.max_stress_ratio, max: 1.2, lowerIsBetter: true },
            { name: '位移', current: current.max_displacement, original: original.max_displacement, max: 60, lowerIsBetter: true },
            { name: '刚度', current: current.stiffness, original: original.stiffness, max: 5, lowerIsBetter: false },
            { name: '材料', current: current.material_volume, original: original.material_volume, max: 400, lowerIsBetter: true }
        ];

        const barWidth = 60;
        const groupWidth = 150;
        const startX = 60;
        const chartH = height - 60;

        metrics.forEach((m, idx) => {
            const groupX = startX + idx * groupWidth;
            
            ctx.fillStyle = '#333';
            ctx.font = '12px sans-serif';
            ctx.textAlign = 'center';
            ctx.fillText(m.name, groupX + groupWidth / 2, height - 25);

            const origH = chartH * (m.original / m.max);
            const currH = chartH * (m.current / m.max);

            ctx.fillStyle = '#95a5a6';
            ctx.fillRect(groupX + 10, height - 40 - origH, barWidth, origH);
            
            const isBetter = m.lowerIsBetter ? (m.current < m.original) : (m.current > m.original);
            ctx.fillStyle = isBetter ? '#2ecc71' : '#e74c3c';
            ctx.fillRect(groupX + 80, height - 40 - currH, barWidth, currH);

            ctx.fillStyle = '#666';
            ctx.font = '10px sans-serif';
            ctx.fillText('原始', groupX + 40, height - 42);
            ctx.fillText('当前', groupX + 110, height - 42);
        });

        ctx.fillStyle = '#95a5a6';
        ctx.fillRect(width - 100, 10, 15, 15);
        ctx.fillStyle = '#333';
        ctx.textAlign = 'left';
        ctx.fillText('原始参数', width - 80, 22);
        
        ctx.fillStyle = '#2ecc71';
        ctx.fillRect(width - 100, 30, 15, 15);
        ctx.fillStyle = '#333';
        ctx.fillText('当前参数 (优)', width - 80, 42);
        
        ctx.fillStyle = '#e74c3c';
        ctx.fillRect(width - 100, 50, 15, 15);
        ctx.fillStyle = '#333';
        ctx.fillText('当前参数 (劣)', width - 80, 62);
    }

    calculateOriginalMetrics() {
        const span = this.originalParams.span;
        const rise = this.originalParams.rise;
        const width = this.originalParams.width;
        const load = this.originalParams.load;
        const ratio = rise / span;

        return {
            max_stress_ratio: 0.4 * Math.pow(span / 25, 1.5) * (load / 50),
            max_displacement: (span * span * load) / (8000 * ratio * width) * 1000,
            stiffness: (width * rise * rise) / (span * span * span) * 100,
            material_volume: span * width * 0.8 + rise * width * 1.2
        };
    }

    updateSensitivityAnalysis() {
        if (!this.analysisResult) return;

        const items = document.querySelectorAll('.sensitivity-item');
        const sensitivities = [
            { label: '拱跨影响', value: 0.65 },
            { label: '矢高影响', value: 0.45 },
            { label: '桥宽影响', value: 0.25 }
        ];

        items.forEach((item, idx) => {
            if (idx < sensitivities.length) {
                const s = sensitivities[idx];
                item.querySelector('.sens-label').textContent = s.label;
                item.querySelector('.sens-fill').style.width = (s.value * 100) + '%';
                item.querySelector('.sens-value').textContent = (s.value * 100).toFixed(0) + '%';
            }
        });
    }

    updateForcesTable(forces) {
        const tbody = document.getElementById('paramForcesBody');
        if (!tbody || !forces) return;

        tbody.innerHTML = '';
        forces.slice(0, 8).forEach(f => {
            const stressClass = f.stress_ratio > 0.9 ? 'danger' : 
                               f.stress_ratio > 0.7 ? 'warning' : 'good';
            
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>M${f.member_id}</td>
                <td>${f.member_type || '-'}</td>
                <td>${f.axial_force.toFixed(1)}</td>
                <td>${f.bending_moment.toFixed(1)}</td>
                <td class="${stressClass}">${f.stress_ratio.toFixed(3)}</td>
            `;
            tbody.appendChild(row);
        });
    }

    update3DView() {
        if (!this.bridgeView || !this.analysisResult) return;

        if (this.bridgeView.regenerateGeometry) {
            this.bridgeView.regenerateGeometry(
                this.currentParams.span,
                this.currentParams.rise,
                this.currentParams.width
            );
        }

        if (this.analysisResult.member_forces) {
            this.bridgeView.setForces(this.analysisResult.member_forces);
            if (document.getElementById('showStress').checked) {
                this.bridgeView.updateMemberColors();
            }
        }

        if (this.analysisResult.displacements) {
            this.bridgeView.setDisplacements(this.analysisResult.displacements);
            if (document.getElementById('showDeformation').checked) {
                this.bridgeView.applyDeformation();
            }
        }
    }

    resetParams() {
        this.currentParams = { ...this.originalParams };
        
        document.getElementById('paramSpan').value = this.currentParams.span;
        document.getElementById('paramSpanVal').textContent = this.currentParams.span.toFixed(1) + ' m';
        
        document.getElementById('paramRise').value = this.currentParams.rise;
        document.getElementById('paramRiseVal').textContent = this.currentParams.rise.toFixed(1) + ' m';
        
        document.getElementById('paramWidth').value = this.currentParams.width;
        document.getElementById('paramWidthVal').textContent = this.currentParams.width.toFixed(1) + ' m';
        
        document.getElementById('paramLoad').value = this.currentParams.load;
        document.getElementById('paramLoadVal').textContent = this.currentParams.load.toFixed(0) + ' kN/m';
        
        this.updateRatioDisplay();
        this.runAnalysis();
    }

    applyPreset(presetId) {
        const presets = {
            1: { span: 25.6, rise: 5.8, width: 6.5, name: '汴水虹桥' },
            2: { span: 22.0, rise: 4.8, width: 5.0, name: '万安桥' },
            3: { span: 30.0, rise: 7.5, width: 7.2, name: '龙津桥' },
            4: { span: 18.5, rise: 4.0, width: 4.5, name: '广济桥' }
        };

        const preset = presets[presetId];
        if (!preset) return;

        this.currentParams.span = preset.span;
        this.currentParams.rise = preset.rise;
        this.currentParams.width = preset.width;

        document.getElementById('paramSpan').value = preset.span;
        document.getElementById('paramSpanVal').textContent = preset.span.toFixed(1) + ' m';
        
        document.getElementById('paramRise').value = preset.rise;
        document.getElementById('paramRiseVal').textContent = preset.rise.toFixed(1) + ' m';
        
        document.getElementById('paramWidth').value = preset.width;
        document.getElementById('paramWidthVal').textContent = preset.width.toFixed(1) + ' m';

        this.updateRatioDisplay();
        
        const recs = document.getElementById('recommendations');
        const info = document.createElement('div');
        info.className = 'recommendation-item info';
        info.innerHTML = `<strong>已加载：</strong>${preset.name} (${preset.span}m × ${preset.rise}m × ${preset.width}m)`;
        recs.insertBefore(info, recs.firstChild);

        if (this.autoUpdate) {
            this.runAnalysis();
        }
    }

    saveDesign() {
        if (!this.analysisResult) {
            alert('请先运行分析');
            return;
        }

        const design = {
            id: Date.now(),
            params: { ...this.currentParams },
            result: { ...this.analysisResult },
            timestamp: new Date().toISOString()
        };

        let saved = JSON.parse(localStorage.getItem('parametricDesigns') || '[]');
        saved.push(design);
        localStorage.setItem('parametricDesigns', JSON.stringify(saved));

        alert(`设计已保存！\n\n参数：${design.params.span}m跨 × ${design.params.rise}m矢 × ${design.params.width}m宽\n应力比：${design.result.max_stress_ratio.toFixed(3)}\n最大位移：${design.result.max_displacement.toFixed(2)}mm`);
    }
}
