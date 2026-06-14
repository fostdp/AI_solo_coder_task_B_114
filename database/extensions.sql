-- =====================================================================
-- 古代木拱桥系统扩展表
-- 包含：动态荷载分析、历史对比、加固优化、参数化设计
-- 新增表，不修改任何现有表，保证向后兼容
-- =====================================================================

-- =====================================================================
-- 1. 动态荷载与疲劳分析表
-- =====================================================================

CREATE TABLE IF NOT EXISTS dynamic_load_analyses (
    analysis_id SERIAL PRIMARY KEY,
    bridge_id INTEGER REFERENCES bridges(bridge_id),
    analysis_type VARCHAR(50) NOT NULL,
    duration_seconds NUMERIC(10,2) NOT NULL,
    crowd_density NUMERIC(6,3) NOT NULL,
    time_step NUMERIC(6,3) NOT NULL,
    random_seed BIGINT,
    max_load NUMERIC(10,2),
    avg_load NUMERIC(10,2),
    total_agents INTEGER,
    load_cycles_per_day NUMERIC(10,2),
    estimated_life_years NUMERIC(12,2),
    total_damage NUMERIC(10,6),
    max_damage_member_id INTEGER,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    parameters JSONB
);

CREATE TABLE IF NOT EXISTS load_spectrum_data (
    id SERIAL PRIMARY KEY,
    analysis_id INTEGER REFERENCES dynamic_load_analyses(analysis_id) ON DELETE CASCADE,
    time_step INTEGER NOT NULL,
    time_seconds NUMERIC(10,3) NOT NULL,
    total_load NUMERIC(10,2) NOT NULL,
    active_agent_count INTEGER,
    load_distribution JSONB
);

CREATE INDEX IF NOT EXISTS idx_load_spectrum_analysis ON load_spectrum_data(analysis_id);

CREATE TABLE IF NOT EXISTS fatigue_member_results (
    id SERIAL PRIMARY KEY,
    analysis_id INTEGER REFERENCES dynamic_load_analyses(analysis_id) ON DELETE CASCADE,
    member_id INTEGER NOT NULL,
    member_type VARCHAR(50),
    cycle_count INTEGER,
    damage_cumulative NUMERIC(12,8),
    remaining_life NUMERIC(12,2),
    fatigue_safety_factor NUMERIC(8,4),
    damage_contribution JSONB,
    stress_range JSONB
);

CREATE INDEX IF NOT EXISTS idx_fatigue_analysis ON fatigue_member_results(analysis_id);

-- =====================================================================
-- 2. 历史桥梁技术对比表
-- =====================================================================

CREATE TABLE IF NOT EXISTS historical_bridge_database (
    id SERIAL PRIMARY KEY,
    bridge_code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    dynasty VARCHAR(20) NOT NULL,
    typology VARCHAR(30) NOT NULL,
    span_length NUMERIC(8,2),
    arch_rise NUMERIC(8,2),
    deck_width NUMERIC(8,2),
    total_length NUMERIC(10,2),
    material_type VARCHAR(50),
    construction_method VARCHAR(100),
    historical_era VARCHAR(100),
    key_innovation TEXT,
    location VARCHAR(100),
    built_year INTEGER,
    destroyed_year INTEGER,
    status VARCHAR(20) DEFAULT 'existing',
    source_references TEXT[],
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_historical_dynasty ON historical_bridge_database(dynasty);
CREATE INDEX IF NOT EXISTS idx_historical_typology ON historical_bridge_database(typology);

CREATE TABLE IF NOT EXISTS comparison_analyses (
    analysis_id SERIAL PRIMARY KEY,
    bridge_a_id INTEGER NOT NULL,
    bridge_b_id INTEGER NOT NULL,
    analysis_name VARCHAR(100),
    metrics_a JSONB,
    metrics_b JSONB,
    normalized_scores JSONB,
    radar_data JSONB,
    advantages_a TEXT[],
    advantages_b TEXT[],
    historical_notes TEXT[],
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    user_notes TEXT
);

CREATE TABLE IF NOT EXISTS tech_evolution_points (
    id SERIAL PRIMARY KEY,
    period VARCHAR(20) NOT NULL,
    year INTEGER,
    innovation VARCHAR(200) NOT NULL,
    impact_score NUMERIC(5,2),
    description TEXT,
    related_bridges INTEGER[],
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================================
-- 3. 加固方案优化表
-- =====================================================================

CREATE TABLE IF NOT EXISTS reinforcement_methods (
    method_code VARCHAR(30) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    base_cost_factor NUMERIC(8,4),
    base_stiffness_gain NUMERIC(8,4),
    base_strength_gain NUMERIC(8,4),
    base_durability_gain NUMERIC(8,4),
    base_complexity NUMERIC(8,4),
    heritage_impact NUMERIC(8,4),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reinforcement_optimizations (
    analysis_id SERIAL PRIMARY KEY,
    bridge_id INTEGER REFERENCES bridges(bridge_id),
    target_nodes INTEGER[],
    population_size INTEGER,
    max_generations INTEGER,
    random_seed BIGINT,
    weight_preferences JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reinforcement_solutions (
    id SERIAL PRIMARY KEY,
    analysis_id INTEGER REFERENCES reinforcement_optimizations(analysis_id) ON DELETE CASCADE,
    solution_rank INTEGER,
    pareto_optimal BOOLEAN DEFAULT FALSE,
    method VARCHAR(30) REFERENCES reinforcement_methods(method_code),
    cfrp_thickness NUMERIC(6,3),
    cfrp_layers INTEGER,
    iron_hoop_count INTEGER,
    iron_hoop_width NUMERIC(6,3),
    steel_plate_thickness NUMERIC(6,3),
    wooden_splice_length NUMERIC(6,3),
    cost_increase_factor NUMERIC(8,4),
    stiffness_gain NUMERIC(8,4),
    strength_gain NUMERIC(8,4),
    durability_gain NUMERIC(8,4),
    construction_complexity NUMERIC(8,4),
    heritage_impact NUMERIC(8,4),
    overall_score NUMERIC(8,4),
    fitness_values JSONB,
    parameters JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_reinforcement_analysis ON reinforcement_solutions(analysis_id);
CREATE INDEX IF NOT EXISTS idx_reinforcement_score ON reinforcement_solutions(overall_score DESC);

-- =====================================================================
-- 4. 参数化设计表
-- =====================================================================

CREATE TABLE IF NOT EXISTS parametric_analyses (
    analysis_id SERIAL PRIMARY KEY,
    bridge_id INTEGER REFERENCES bridges(bridge_id),
    span_length NUMERIC(8,2) NOT NULL,
    arch_rise NUMERIC(8,2) NOT NULL,
    deck_width NUMERIC(8,2) NOT NULL,
    rise_span_ratio NUMERIC(6,4),
    load_value NUMERIC(10,2),
    member_count INTEGER,
    node_count INTEGER,
    max_stress_ratio NUMERIC(8,4),
    max_displacement NUMERIC(8,4),
    max_displacement_node INTEGER,
    total_volume NUMERIC(12,4),
    material_efficiency NUMERIC(10,4),
    analysis_duration_ms INTEGER,
    valid BOOLEAN DEFAULT TRUE,
    validation_message VARCHAR(200),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    member_forces JSONB,
    displacements JSONB,
    yingzao_comparison JSONB
);

CREATE INDEX IF NOT EXISTS idx_parametric_bridge ON parametric_analyses(bridge_id);
CREATE INDEX IF NOT EXISTS idx_parametric_ratio ON parametric_analyses(rise_span_ratio);

CREATE TABLE IF NOT EXISTS parametric_design_sessions (
    session_id SERIAL PRIMARY KEY,
    bridge_id INTEGER REFERENCES bridges(bridge_id),
    user_id VARCHAR(50),
    started_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    design_points INTEGER DEFAULT 0,
    best_design_id INTEGER REFERENCES parametric_analyses(analysis_id),
    notes TEXT
);

-- =====================================================================
-- 插入基础数据
-- =====================================================================

-- 历史桥梁基础数据
INSERT INTO historical_bridge_database 
    (bridge_code, name, dynasty, typology, span_length, arch_rise, deck_width, 
     total_length, material_type, construction_method, historical_era, key_innovation,
     location, built_year, status) VALUES
    ('HB-001', '灞桥', 'han_jin', 'beam_bridge', 18.0, 2.5, 4.5, 400.0, 
     '木梁石墩', '石墩木梁', '汉晋时期(公元前206年-公元420年)', '多跨简支木梁桥，石砌桥墩',
     '陕西西安', -100, 'destroyed'),
    ('HB-002', '枫桥', 'tang', 'beam_bridge', 18.5, 3.8, 4.2, 24.0,
     '木材', '单孔木拱', '唐代(公元618年-907年)', '木拱技术萌芽，向拱结构过渡',
     '江苏苏州', 750, 'reconstructed'),
    ('HB-003', '汴水虹桥', 'song', 'through_arch', 25.6, 5.8, 6.5, 32.0,
     '木材', '叠梁拱/贯木拱', '北宋(公元960年-1127年)', '贯木拱技术成熟，无柱大跨木拱桥',
     '河南开封', 1050, 'destroyed'),
    ('HB-004', '龙津桥', 'ming', 'gallery_bridge', 28.5, 6.2, 5.2, 35.0,
     '木材', '木拱廊桥', '明代(公元1368年-1644年)', '木拱廊桥，廊屋保护结构',
     '福建寿宁', 1450, 'existing'),
    ('HB-005', '千乘桥', 'ming', 'timber_arch', 27.3, 5.9, 5.0, 34.0,
     '木材', '贯木拱', '明代', '三节拱五节拱组合技术',
     '福建屏南', 1520, 'existing'),
    ('HB-006', '飞虹桥', 'qing', 'gallery_bridge', 19.5, 4.2, 4.5, 26.0,
     '木材', '木拱廊桥', '清代(公元1644年-1912年)', '工艺精细化与装饰艺术发展',
     '浙江泰顺', 1750, 'existing'),
    ('HB-007', '安澜桥', 'song', 'beam_bridge', 24.0, 5.2, 4.8, 31.0,
     '竹木', '竹索木桥', '宋代', '竹索加固木梁技术',
     '四川都江堰', 1000, 'reconstructed')
ON CONFLICT (bridge_code) DO NOTHING;

-- 加固方法基础数据
INSERT INTO reinforcement_methods 
    (method_code, name, description, base_cost_factor, base_stiffness_gain,
     base_strength_gain, base_durability_gain, base_complexity, heritage_impact) VALUES
    ('iron_hoop', '传统铁箍加固', '传统铁箍加固，历史真实性较好', 1.2, 0.15, 0.20, 0.25, 0.3, 0.4),
    ('cfrp', '碳纤维布(CFRP)加固', '碳纤维布加固，高效且隐蔽', 2.5, 0.35, 0.40, 0.45, 0.6, 0.2),
    ('steel_plate', '钢板粘贴加固', '钢板粘贴加固，刚度提升显著', 1.8, 0.30, 0.35, 0.30, 0.5, 0.5),
    ('wooden_splice', '木榫拼接加固', '木榫拼接加固，历史真实性最佳', 1.5, 0.10, 0.15, 0.20, 0.4, 0.1),
    ('combined', '组合加固方案', '铁箍+CFRP组合加固，综合性能最优', 3.0, 0.50, 0.55, 0.50, 0.8, 0.3)
ON CONFLICT (method_code) DO NOTHING;

-- 技术演进数据
INSERT INTO tech_evolution_points 
    (period, year, innovation, impact_score, description, related_bridges) VALUES
    ('西周', -1000, '简支木梁桥出现', 30.0, '最早的木结构桥梁形式，结构简单', ARRAY[]::INTEGER[]),
    ('秦汉', -200, '石墩木梁桥普及', 45.0, '石砌桥墩技术成熟，实现较大跨越能力', ARRAY[]::INTEGER[]),
    ('南北朝', 500, '木拱技术萌芽', 60.0, '开始探索拱形结构，受力更合理', ARRAY[]::INTEGER[]),
    ('唐代', 700, '单孔木拱桥出现', 70.0, '木拱技术逐渐成熟，跨度能力提升', ARRAY[]::INTEGER[]),
    ('北宋', 1050, '贯木拱技术成熟', 95.0, '叠梁拱与贯木拱技术达到顶峰，实现无柱大跨', ARRAY[3]::INTEGER[]),
    ('南宋', 1200, '木拱廊桥发展', 85.0, '增加廊屋保护结构，延长使用寿命', ARRAY[]::INTEGER[]),
    ('明代', 1450, '三节拱五节拱组合', 88.0, '多节拱组合技术，跨度进一步突破', ARRAY[4,5]::INTEGER[]),
    ('清代', 1700, '工艺精细化与装饰', 80.0, '工艺技术成熟，装饰艺术发展', ARRAY[6]::INTEGER[])
ON CONFLICT DO NOTHING;

-- =====================================================================
-- 新功能完成标记
-- =====================================================================

COMMENT ON TABLE dynamic_load_analyses IS '新增：动态荷载分析表（2026-06新增）';
COMMENT ON TABLE historical_bridge_database IS '新增：历史桥梁数据库（2026-06新增）';
COMMENT ON TABLE reinforcement_optimizations IS '新增：加固优化分析表（2026-06新增）';
COMMENT ON TABLE parametric_analyses IS '新增：参数化设计分析表（2026-06新增）';
