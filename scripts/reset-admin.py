#!/usr/bin/env python3

import os
import sys
import getpass
import bcrypt
import psycopg2
from datetime import datetime
import re

def main():
    print("🔑 CloudBox Admin Password Reset Tool (Python)")
    print("==============================================")
    print()
    
    # Get database connection
    db_url = os.getenv('DATABASE_URL') or os.getenv('DB_CONNECTION_STRING')
    if not db_url:
        db_url = "postgres://cloudbox:cloudbox_dev_password@localhost:5432/cloudbox"
        print(f"Using default database URL: {db_url}")
        input("Press Enter to continue or Ctrl+C to abort...")
    
    # Parse database URL
    try:
        # Parse postgres://user:pass@host:port/dbname
        import urllib.parse
        parsed = urllib.parse.urlparse(db_url)
        
        conn_params = {
            'host': parsed.hostname,
            'port': parsed.port or 5432,
            'database': parsed.path[1:],  # Remove leading slash
            'user': parsed.username,
            'password': parsed.password
        }
        
        print("📡 Connecting to database...")
        conn = psycopg2.connect(**conn_params)
        cursor = conn.cursor()
        print("✅ Connected to database successfully")
        print()
        
    except Exception as e:
        print(f"❌ Failed to connect to database: {e}")
        sys.exit(1)
    
    # Get user input
    email = get_email_input()
    password = get_password_input()
    
    # Confirm action
    if not confirm_action(email):
        print("❌ Operation cancelled")
        return
    
    # Hash password
    print("🔐 Hashing password...")
    hashed_password = bcrypt.hashpw(password.encode('utf-8'), bcrypt.gensalt())
    
    try:
        # Check if user exists
        cursor.execute("SELECT id, name, email, role, is_active FROM users WHERE email = %s", (email,))
        user = cursor.fetchone()
        
        if user is None:
            # Create new admin user
            print(f"👤 User {email} not found. Creating new admin user...")
            name = get_name_input(email)
            
            cursor.execute("""
                INSERT INTO users (created_at, updated_at, email, name, password_hash, role, is_active) 
                VALUES (%s, %s, %s, %s, %s, %s, %s) 
                RETURNING id
            """, (
                datetime.now(),
                datetime.now(),
                email,
                name,
                hashed_password.decode('utf-8'),
                'admin',
                True
            ))
            
            user_id = cursor.fetchone()[0]
            conn.commit()
            
            print(f"✅ Admin user '{email}' created successfully")
            print(f"🆔 User ID: {user_id}")
            
        else:
            # Update existing user
            user_id, user_name, user_email, user_role, user_active = user
            print(f"👤 Found existing user: {user_name} ({user_email})")
            
            cursor.execute("""
                UPDATE users 
                SET password_hash = %s, role = %s, is_active = %s, updated_at = %s 
                WHERE email = %s
            """, (
                hashed_password.decode('utf-8'),
                'admin',
                True,
                datetime.now(),
                email
            ))
            
            conn.commit()
            
            print(f"✅ Password reset successfully for user '{email}'")
            print("✅ User role set to 'admin'")
            print("✅ User account activated")
        
        # Get final user state
        cursor.execute("SELECT id, name, email, role, is_active FROM users WHERE email = %s", (email,))
        final_user = cursor.fetchone()
        
        print()
        print("🎉 Operation completed successfully!")
        print(f"📧 Email: {final_user[2]}")
        print(f"👤 Name: {final_user[1]}")
        print(f"🛡️  Role: {final_user[3]}")
        print(f"✅ Active: {final_user[4]}")
        print(f"🆔 User ID: {final_user[0]}")
        print()
        print("You can now login to CloudBox with these credentials.")
        
    except Exception as e:
        conn.rollback()
        print(f"❌ Database error: {e}")
        sys.exit(1)
    
    finally:
        cursor.close()
        conn.close()

def get_email_input():
    while True:
        email = input("📧 Enter admin email address: ").strip()
        
        if not email:
            print("❌ Email cannot be empty")
            continue
            
        if not re.match(r'^[^@]+@[^@]+\.[^@]+$', email):
            print("❌ Please enter a valid email address")
            continue
            
        return email

def get_password_input():
    while True:
        password = getpass.getpass("🔑 Enter new password: ")
        
        if len(password) < 6:
            print("❌ Password must be at least 6 characters long")
            continue
            
        confirm_password = getpass.getpass("🔑 Confirm new password: ")
        
        if password != confirm_password:
            print("❌ Passwords do not match")
            continue
            
        return password

def get_name_input(email):
    # Extract name from email as default
    default_name = email.split('@')[0].replace('.', ' ').title()
    
    name = input(f"👤 Enter full name (default: {default_name}): ").strip()
    
    return name if name else default_name

def confirm_action(email):
    print(f"⚠️  This will reset the password for '{email}' and set role to 'admin'")
    response = input("Are you sure you want to continue? (yes/no): ").strip().lower()
    
    return response in ['yes', 'y']

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n❌ Operation cancelled by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n❌ Unexpected error: {e}")
        sys.exit(1)