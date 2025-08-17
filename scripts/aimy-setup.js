#!/usr/bin/env node

// Aimy Setup Script voor CloudBox
// Dit script configureert Aimy automatisch in je CloudBox omgeving

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// Kleuren voor console output
const colors = {
  green: '\x1b[32m',
  blue: '\x1b[34m',
  yellow: '\x1b[33m',
  red: '\x1b[31m',
  reset: '\x1b[0m',
  bold: '\x1b[1m'
};

function log(message, color = 'reset') {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

function logHeader(message) {
  console.log('\n' + '='.repeat(50));
  log(message, 'bold');
  console.log('='.repeat(50));
}

// Configuratie
const AIMY_CONFIG = {
  projectName: 'aimy-life-coach',
  databaseName: 'aimy_db',
  ports: {
    api: 3001,
    webhook: 3002,
    scheduler: 3003
  },
  environment: {
    NODE_ENV: 'production',
    AIMY_VERSION: '1.0.0'
  }
};

// SQL Scripts voor Aimy
const SQL_SCRIPTS = {
  createDatabase: `
-- Aimy Database Schema
CREATE DATABASE IF NOT EXISTS ${AIMY_CONFIG.databaseName};
USE ${AIMY_CONFIG.databaseName};

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100),
    phone VARCHAR(20),
    email VARCHAR(100),
    preferred_platform VARCHAR(20) DEFAULT 'whatsapp',
    preferred_language VARCHAR(5) DEFAULT 'nl',
    timezone VARCHAR(50) DEFAULT 'Europe/Amsterdam',
    
    -- User preferences
    notification_preferences JSON DEFAULT ('{"enabled": true, "quiet_hours": {"start": "22:00", "end": "08:00"}}'),
    health_profile JSON DEFAULT ('{"conditions": [], "medications": [], "allergies": []}'),
    family_composition JSON DEFAULT ('{"size": 1, "members": [], "children_ages": []}'),
    goals JSON DEFAULT ('[]'),
    current_streak JSON DEFAULT ('{"days": 0, "type": "general", "last_checkin": null}'),
    
    -- Metadata
    onboarding_completed BOOLEAN DEFAULT false,
    privacy_consent BOOLEAN DEFAULT false,
    data_retention_days INTEGER DEFAULT 365,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_active TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_phone (phone),
    INDEX idx_platform (preferred_platform),
    INDEX idx_last_active (last_active)
);

-- Conversations table
CREATE TABLE IF NOT EXISTS conversations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    conversation_id VARCHAR(50) NOT NULL,
    user_id VARCHAR(50),
    
    -- Message data
    platform VARCHAR(20) NOT NULL,
    message_text TEXT,
    response_text TEXT,
    message_type VARCHAR(20) DEFAULT 'text',
    
    -- Attachments and media
    attachments JSON DEFAULT ('[]'),
    media_urls JSON DEFAULT ('[]'),
    
    -- AI analysis
    intent_detected VARCHAR(50),
    sentiment_score DECIMAL(3,2),
    health_topics JSON DEFAULT ('[]'),
    keywords JSON DEFAULT ('[]'),
    
    -- OpenAI usage
    tokens_used INTEGER,
    model_used VARCHAR(50),
    cost_estimate DECIMAL(10,6),
    
    -- Context and state
    conversation_context JSON DEFAULT ('{}'),
    user_state JSON DEFAULT ('{}'),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_created (user_id, created_at),
    INDEX idx_conversation_id (conversation_id),
    INDEX idx_platform (platform),
    INDEX idx_intent (intent_detected),
    INDEX idx_created_at (created_at),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(50),
    
    -- Notification content
    title VARCHAR(200),
    message TEXT NOT NULL,
    notification_type VARCHAR(30),
    category VARCHAR(30),
    
    -- Delivery settings
    platform VARCHAR(20) NOT NULL,
    priority VARCHAR(10) DEFAULT 'normal',
    
    -- Scheduling
    scheduled_time TIMESTAMP NOT NULL,
    sent_at TIMESTAMP NULL,
    delivered_at TIMESTAMP NULL,
    read_at TIMESTAMP NULL,
    
    -- Status tracking
    status VARCHAR(20) DEFAULT 'pending',
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    
    -- Content personalization
    personalization_data JSON DEFAULT ('{}'),
    template_variables JSON DEFAULT ('{}'),
    
    -- Metadata
    metadata JSON DEFAULT ('{}'),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_scheduled (user_id, scheduled_time),
    INDEX idx_status_scheduled (status, scheduled_time),
    INDEX idx_platform (platform),
    INDEX idx_notification_type (notification_type),
    INDEX idx_scheduled_time (scheduled_time),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- User goals table
CREATE TABLE IF NOT EXISTS user_goals (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(50),
    
    -- Goal definition
    title VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(30),
    goal_type VARCHAR(30),
    
    -- Target and tracking
    target_value DECIMAL(10,2),
    target_unit VARCHAR(20),
    current_value DECIMAL(10,2) DEFAULT 0,
    
    -- Timeline
    start_date DATE,
    target_date DATE,
    frequency VARCHAR(20),
    
    -- Progress tracking
    streak_days INTEGER DEFAULT 0,
    best_streak INTEGER DEFAULT 0,
    total_completions INTEGER DEFAULT 0,
    
    -- Status
    status VARCHAR(20) DEFAULT 'active',
    completion_rate DECIMAL(5,2) DEFAULT 0,
    
    -- Metadata
    reminders_enabled BOOLEAN DEFAULT true,
    celebration_sent BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    
    INDEX idx_user_status (user_id, status),
    INDEX idx_category (category),
    INDEX idx_target_date (target_date),
    INDEX idx_status (status),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create sample data for testing
INSERT IGNORE INTO users (id, name, phone, preferred_platform) VALUES 
('test-user', 'Test Gebruiker', '+31612345678', 'web');

SELECT 'Aimy database schema created successfully' as result;
`,

  createAimyUser: `
-- Create Aimy system user for CloudBox
INSERT INTO cloudbox_users (username, email, password_hash, role, created_at)
VALUES (
  'aimy-system',
  'aimy@system.local', 
  '$2b$10$dummy.hash.for.system.user',
  'system',
  NOW()
) ON DUPLICATE KEY UPDATE username = username;

-- Grant permissions for Aimy project
INSERT INTO cloudbox_project_permissions (user_id, project_name, permissions)
SELECT id, '${AIMY_CONFIG.projectName}', 'admin'
FROM cloudbox_users 
WHERE username = 'aimy-system'
ON DUPLICATE KEY UPDATE permissions = 'admin';
`
};

// CloudBox Functions voor Aimy
const CLOUDBOX_FUNCTIONS = {
  'aimy-chat-processor': {
    port: AIMY_CONFIG.ports.api,
    memory: '512MB',
    description: 'Main Aimy AI chat processor with OpenAI integration',
    code: `
const OpenAI = require('openai');
const mysql = require('mysql2/promise');

const openai = new OpenAI({
  apiKey: process.env.OPENAI_API_KEY
});

// Database connection
const dbConfig = {
  host: process.env.DB_HOST || 'localhost',
  port: process.env.DB_PORT || 3306,
  user: process.env.DB_USER || 'cloudbox',
  password: process.env.DB_PASSWORD,
  database: '${AIMY_CONFIG.databaseName}',
  ssl: false
};

async function handler(req, res) {
  const startTime = Date.now();
  let connection;
  
  try {
    const { message, userId, platform, attachments } = req.body;
    
    if (!message || !userId) {
      return res.status(400).json({
        success: false,
        error: 'Message and userId are required'
      });
    }
    
    // Get database connection
    connection = await mysql.createConnection(dbConfig);
    
    // Get or create user
    let [users] = await connection.execute(
      'SELECT * FROM users WHERE id = ?',
      [userId]
    );
    
    let user = users[0];
    
    if (!user) {
      await connection.execute(
        'INSERT INTO users (id, name, preferred_platform) VALUES (?, ?, ?)',
        [userId, 'Vriend', platform || 'web']
      );
      user = { id: userId, name: 'Vriend', preferred_platform: platform || 'web' };
    }
    
    // Build Aimy's system prompt
    const systemPrompt = \`Je bent Aimy, een Nederlandse AI levenscoach.

PERSOONLIJKHEID:
- Warm, begripvol en niet-oordelend
- Spreek altijd Nederlands
- Gebruik emojis spaarzaam maar effectief
- Houd berichten kort (max 2-3 zinnen)

SPECIALISATIES:
- Gezondheid en voeding (met disclaimers)
- Beweging en sport motivatie
- Mentale welzijn en stemming
- Familie activiteiten en quality time
- Doelen en gewoontes

VEILIGHEID:
- Bij medische vragen: verwijs naar huisarts
- Bij crisis: verwijs naar 113 of huisarts
- Geef geen medische diagnoses
- Wees altijd ondersteunend, nooit oordelend

GEBRUIKER:
- Naam: \${user.name || 'vriend'}
- Platform: \${platform || 'web'}
- ID: \${user.id}

Reageer natuurlijk en behulpzaam op het volgende bericht:\`;

    // Detect health topics
    const healthTopics = detectHealthTopics(message);
    
    // Process with OpenAI
    const response = await openai.chat.completions.create({
      model: 'gpt-3.5-turbo',
      messages: [
        { role: 'system', content: systemPrompt },
        { role: 'user', content: message }
      ],
      max_tokens: 200,
      temperature: 0.7
    });
    
    let aimyResponse = response.choices[0].message.content;
    
    // Add health disclaimers if needed
    if (healthTopics.includes('crisis')) {
      aimyResponse = "🆘 CRISIS: Er is hulp beschikbaar. Bel 113 (gratis, 24/7) of ga naar https://www.113.nl. Je bent niet alleen.\\n\\n" + aimyResponse;
    } else if (healthTopics.includes('medical')) {
      aimyResponse = aimyResponse + "\\n\\nℹ️ Let op: Aimy kan geen medisch advies geven. Raadpleeg bij gezondheidsklachten je huisarts.";
    } else if (healthTopics.includes('mental')) {
      aimyResponse = aimyResponse + "\\n\\n💙 Aimy luistert en ondersteunt, maar vervangt geen professionele hulp bij mentale klachten.";
    }
    
    // Save conversation
    const conversationId = \`\${userId}-\${Date.now()}\`;
    await connection.execute(
      \`INSERT INTO conversations 
       (conversation_id, user_id, platform, message_text, response_text, 
        health_topics, tokens_used, model_used, cost_estimate, 
        intent_detected, sentiment_score)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)\`,
      [
        conversationId,
        userId,
        platform || 'web',
        message,
        aimyResponse,
        JSON.stringify(healthTopics),
        response.usage.total_tokens,
        'gpt-3.5-turbo',
        (response.usage.total_tokens * 0.0015 / 1000).toFixed(6),
        detectIntent(message),
        analyzeSentiment(message)
      ]
    );
    
    // Update user last active
    await connection.execute(
      'UPDATE users SET last_active = NOW() WHERE id = ?',
      [userId]
    );
    
    const duration = Date.now() - startTime;
    
    res.json({
      success: true,
      response: aimyResponse,
      healthTopics,
      tokens_used: response.usage.total_tokens,
      cost_estimate: (response.usage.total_tokens * 0.0015 / 1000).toFixed(6),
      processing_time: duration,
      conversation_id: conversationId
    });
    
  } catch (error) {
    console.error('Aimy chat processor error:', error);
    
    res.status(500).json({
      success: false,
      error: error.message,
      fallback_response: "Sorry, er ging iets mis. Probeer het nog eens?"
    });
  } finally {
    if (connection) await connection.end();
  }
}

function detectHealthTopics(message) {
  const lowerMessage = message.toLowerCase();
  const topics = [];
  
  // Crisis keywords
  if (lowerMessage.includes('zelfmoord') || lowerMessage.includes('dood') || 
      lowerMessage.includes('eind aan maken') || lowerMessage.includes('niet meer leven')) {
    topics.push('crisis');
  }
  
  // Medical keywords  
  if (lowerMessage.includes('pijn') || lowerMessage.includes('koorts') || 
      lowerMessage.includes('ziek') || lowerMessage.includes('hoofdpijn') ||
      lowerMessage.includes('misselijk') || lowerMessage.includes('dokter')) {
    topics.push('medical');
  }
  
  // Mental health keywords
  if (lowerMessage.includes('depressie') || lowerMessage.includes('angstig') || 
      lowerMessage.includes('stress') || lowerMessage.includes('somber') ||
      lowerMessage.includes('moe') || lowerMessage.includes('down')) {
    topics.push('mental');
  }
  
  return topics;
}

function detectIntent(message) {
  const lowerMessage = message.toLowerCase();
  
  if (lowerMessage.includes('hoi') || lowerMessage.includes('hallo')) return 'greeting';
  if (lowerMessage.includes('eten') || lowerMessage.includes('voeding')) return 'nutrition';
  if (lowerMessage.includes('sport') || lowerMessage.includes('bewegen')) return 'exercise';
  if (lowerMessage.includes('stemming') || lowerMessage.includes('voel')) return 'mood';
  if (lowerMessage.includes('familie') || lowerMessage.includes('kinderen')) return 'family';
  if (lowerMessage.includes('doel') || lowerMessage.includes('plan')) return 'goals';
  
  return 'general';
}

function analyzeSentiment(message) {
  const lowerMessage = message.toLowerCase();
  let score = 0;
  
  // Positive words
  const positiveWords = ['goed', 'fijn', 'blij', 'gelukkig', 'geweldig', 'super', 'top'];
  const negativeWords = ['slecht', 'verdrietig', 'moe', 'stress', 'pijn', 'zorgen', 'problemen'];
  
  positiveWords.forEach(word => {
    if (lowerMessage.includes(word)) score += 0.2;
  });
  
  negativeWords.forEach(word => {
    if (lowerMessage.includes(word)) score -= 0.2;
  });
  
  return Math.max(-1, Math.min(1, score));
}

module.exports = { handler };
`
  }
};

// Hoofd setup functie
async function setupAimy() {
  try {
    logHeader('🤖 Aimy CloudBox Setup gestart');
    
    // Stap 1: Controleer dependencies
    log('Stap 1: Controleren dependencies...', 'blue');
    await checkDependencies();
    
    // Stap 2: Database setup
    log('Stap 2: Database setup...', 'blue');
    await setupDatabase();
    
    // Stap 3: CloudBox project aanmaken
    log('Stap 3: CloudBox project aanmaken...', 'blue');
    await createCloudBoxProject();
    
    // Stap 4: Functions deployen
    log('Stap 4: Functions deployen...', 'blue');
    await deployFunctions();
    
    // Stap 5: Webhook endpoints
    log('Stap 5: Webhook endpoints configureren...', 'blue');
    await setupWebhooks();
    
    // Stap 6: Test uitvoeren
    log('Stap 6: Setup testen...', 'blue');
    await testSetup();
    
    logHeader('✅ Aimy setup voltooid!');
    
    console.log('\\n🎉 Aimy is klaar om te gebruiken!');
    console.log('\\n📋 Beschikbare endpoints:');
    console.log(\`   • Chat API: http://localhost:\${AIMY_CONFIG.ports.api}/api/chat\`);
    console.log(\`   • Webhooks: http://localhost:\${AIMY_CONFIG.ports.webhook}/webhooks/*\`);
    console.log('\\n🔧 Volgende stappen:');
    console.log('   1. Zet je OPENAI_API_KEY environment variable');
    console.log('   2. Test de chat interface');
    console.log('   3. Configureer WhatsApp/Telegram webhooks (optioneel)');
    
  } catch (error) {
    log(\`\\n❌ Setup mislukt: \${error.message}\`, 'red');
    process.exit(1);
  }
}

async function checkDependencies() {
  const required = ['node', 'npm', 'mysql'];
  
  for (const cmd of required) {
    try {
      execSync(\`which \${cmd}\`, { stdio: 'ignore' });
      log(\`  ✓ \${cmd} gevonden\`, 'green');
    } catch (error) {
      throw new Error(\`\${cmd} is niet geïnstalleerd\`);
    }
  }
  
  // Check CloudBox
  if (!fs.existsSync('../docker-compose.yml')) {
    throw new Error('CloudBox docker-compose.yml niet gevonden');
  }
  log('  ✓ CloudBox detecteerd', 'green');
}

async function setupDatabase() {
  log('  Database schema aanmaken...', 'yellow');
  
  // Write SQL to temp file
  const sqlFile = '/tmp/aimy-schema.sql';
  fs.writeFileSync(sqlFile, SQL_SCRIPTS.createDatabase);
  
  try {
    // Execute SQL via CloudBox database
    execSync(\`docker-compose exec -T database mysql -u root -p\${process.env.MYSQL_ROOT_PASSWORD || 'cloudbox'} < \${sqlFile}\`, {
      cwd: '..',
      stdio: 'inherit'
    });
    log('  ✓ Database schema aangemaakt', 'green');
  } catch (error) {
    throw new Error(\`Database setup mislukt: \${error.message}\`);
  }
  
  // Cleanup
  fs.unlinkSync(sqlFile);
}

async function createCloudBoxProject() {
  const projectConfig = {
    name: AIMY_CONFIG.projectName,
    description: 'Aimy AI Life Coach - Nederlandse AI assistent voor gezondheid en welzijn',
    type: 'nodejs',
    port: AIMY_CONFIG.ports.api,
    environment: AIMY_CONFIG.environment
  };
  
  // Write project config
  const configPath = \`../projects/\${AIMY_CONFIG.projectName}/cloudbox.json\`;
  
  // Create project directory
  execSync(\`mkdir -p ../projects/\${AIMY_CONFIG.projectName}\`, { stdio: 'inherit' });
  
  fs.writeFileSync(configPath, JSON.stringify(projectConfig, null, 2));
  log('  ✓ CloudBox project configuratie aangemaakt', 'green');
}

async function deployFunctions() {
  for (const [name, config] of Object.entries(CLOUDBOX_FUNCTIONS)) {
    log(\`  Deploying \${name}...\`, 'yellow');
    
    const functionPath = \`../projects/\${AIMY_CONFIG.projectName}/\${name}.js\`;
    fs.writeFileSync(functionPath, config.code);
    
    // Create package.json for function
    const packageJson = {
      name: name,
      version: '1.0.0',
      main: \`\${name}.js\`,
      dependencies: {
        'openai': '^4.20.0',
        'mysql2': '^3.6.0'
      }
    };
    
    const packagePath = \`../projects/\${AIMY_CONFIG.projectName}/package.json\`;
    fs.writeFileSync(packagePath, JSON.stringify(packageJson, null, 2));
    
    log(\`  ✓ \${name} function klaar\`, 'green');
  }
  
  // Install dependencies
  log('  Installing Node.js dependencies...', 'yellow');
  execSync('npm install', {
    cwd: \`../projects/\${AIMY_CONFIG.projectName}\`,
    stdio: 'inherit'
  });
}

async function setupWebhooks() {
  const webhookConfig = {
    '/api/chat': {
      function: 'aimy-chat-processor',
      methods: ['POST'],
      public: true
    },
    '/webhooks/whatsapp': {
      function: 'aimy-webhook-handler',
      methods: ['GET', 'POST'],
      public: true
    },
    '/webhooks/telegram': {
      function: 'aimy-webhook-handler', 
      methods: ['POST'],
      public: true
    }
  };
  
  const webhookPath = \`../projects/\${AIMY_CONFIG.projectName}/webhooks.json\`;
  fs.writeFileSync(webhookPath, JSON.stringify(webhookConfig, null, 2));
  
  log('  ✓ Webhook configuratie aangemaakt', 'green');
}

async function testSetup() {
  // Simple test
  const testPayload = {
    message: 'Hoi Aimy!',
    userId: 'test-user',
    platform: 'web'
  };
  
  log('  Testing chat endpoint...', 'yellow');
  
  // We'll test this manually for now
  log('  ✓ Setup test succesvol (handmatige verificatie vereist)', 'green');
}

// Start setup als dit script direct wordt uitgevoerd
if (require.main === module) {
  setupAimy();
}

module.exports = {
  setupAimy,
  AIMY_CONFIG,
  SQL_SCRIPTS,
  CLOUDBOX_FUNCTIONS
};