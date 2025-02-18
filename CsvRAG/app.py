import streamlit as st
import pandas as pd
import sqlite3
import logging
from itertools import permutations
import json
from datetime import datetime

# **🔹 Set Up Logging**
logging.basicConfig(filename="app.log", level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s")
logger = logging.getLogger()

# **🔹 SQLite Connection and Setup**
def init_db():
    conn = sqlite3.connect("matches.db", check_same_thread=False)
    c = conn.cursor()
    c.execute('''CREATE TABLE IF NOT EXISTS matches
                 (id INTEGER PRIMARY KEY AUTOINCREMENT,
                  donation_id TEXT,
                  mutation_id TEXT,
                  donation_amount INTEGER,
                  unique_code INTEGER,
                  credit REAL,
                  donor_name TEXT,
                  bank_description TEXT,
                  match_type TEXT,
                  match_rules TEXT,
                  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)''')
    conn.commit()
    return conn

conn = init_db()

# **🔹 Function to Validate Required Columns**
def validate_columns(df, required_columns, file_name):
    missing_columns = required_columns - set(df.columns)
    if missing_columns:
        error_msg = f"⚠️ {file_name} is missing columns: {missing_columns}"
        logger.error(error_msg)
        st.error(error_msg)
        st.stop()

# **🔹 Function to Generate Unique Code Variants**
def generate_unique_code_variants(amount, unique_code, is_permutation):
    try:
        # Convert inputs to integers
        amount = int(amount)
        unique_code = int(unique_code)
        
        # Calculate the actual combined value
        actual_value = amount + unique_code
        
        if not is_permutation:
            return [actual_value]
            
        # For permutation match
        results = set([actual_value])  # Start with the actual value
        
        # Get the last two digits of the actual value
        actual_str = str(actual_value)
        base = actual_str[:-2]  # Everything except last 2 digits
        last_two = actual_str[-2:]  # Last 2 digits
        
        # Generate permutations of the last two digits
        for p in permutations(last_two):
            perm_value = int(base + ''.join(p))
            results.add(perm_value)
            
        return list(results)
    except (ValueError, TypeError):
        return [amount]

# **🔹 Function to Process DataFrames**
def preprocess_data(df1, df2, is_permutation, transaction_type):
    # Define required columns based on transaction type
    if transaction_type == "Donation":
        required_columns_1 = {"id", "amount", "unique_code", "full_name", "name", "created"}
    else:  # Top-up
        required_columns_1 = {"id", "amount", "unique_code", "full_name", "name", "created_at"}
    
    required_columns_2 = {"id", "credit", "description", "transfer_time"}

    validate_columns(df1, required_columns_1, "CSV File 1")
    validate_columns(df2, required_columns_2, "CSV File 2")

    # Handle the created/created_at column
    if transaction_type == "Donation":
        df1["created"] = pd.to_datetime(df1["created"], errors="coerce")
    else:
        df1.rename(columns={"created_at": "created"}, inplace=True)
        df1["created"] = pd.to_datetime(df1["created"], errors="coerce")
    
    df2["transfer_time"] = pd.to_datetime(df2["transfer_time"], errors="coerce")

    df1.dropna(subset=["created"], inplace=True)
    df2.dropna(subset=["transfer_time"], inplace=True)

    df1.rename(columns={"id": "csv1_id"}, inplace=True)
    df2.rename(columns={"id": "mutation_id"}, inplace=True)

    df2["amount"] = df2["credit"].astype(float).astype(int)
    df1["unique_code_variants"] = df1.apply(
        lambda row: generate_unique_code_variants(int(row["amount"]), row["unique_code"], is_permutation), axis=1)

    return df1, df2

# **🔹 Function to Check Name Match**
def is_partial_match(full_name, description):
    if pd.isna(full_name) or pd.isna(description):
        return False

    name_parts = full_name.lower().split()
    description_lower = description.lower()

    return any(part in description_lower for part in name_parts)


# **🔹 Function to Perform Matching**
def perform_matching(df1, df2, selected_rules, transaction_type):
    df1_exploded = df1.explode("unique_code_variants")
    df1_exploded.rename(columns={"unique_code_variants": "amount_unique_code"}, inplace=True)

    # Always do amount matching
    matched_df = df1_exploded.merge(df2, how="inner", left_on="amount_unique_code", right_on="credit")

    if matched_df.empty:
        st.warning("⚠️ No matches found after applying Amount Match.")
        return None

    if "Name Match" in selected_rules:
        matched_df = matched_df[
            matched_df.apply(lambda row: is_partial_match(row.get("full_name", ""), row.get("description", "")), axis=1)
        ]

        if matched_df.empty:
            st.warning("⚠️ No matches found after applying Name Match.")
            return None

    if "Time Match" in selected_rules:
        matched_df = matched_df[
            (matched_df["created"] - matched_df["transfer_time"]).abs() <= pd.Timedelta(hours=24)
        ]

        if matched_df.empty:
            st.warning("⚠️ No matches found after applying Time Match.")
            return None

    if transaction_type == "Donation":
        matched_df.rename(columns={"csv1_id": "donation_id", "amount_x": "donation_amount"}, inplace=True)
        final_columns = ["donation_id", "mutation_id", "full_name", "description", "donation_amount", "unique_code", "credit"]
    else:
        matched_df.rename(columns={"csv1_id": "wallet_transaction_id", "amount_x": "topup_amount"}, inplace=True)
        final_columns = ["wallet_transaction_id", "mutation_id", "full_name", "description", "topup_amount", "unique_code", "credit"]

    # Store matches
    match_type = "Permutation Match" if "Permutation Match" in selected_rules else "Exact Match"
    
    for _, row in matched_df[final_columns].iterrows():
        store_match(row, match_type, selected_rules)

    return matched_df[final_columns]

def store_match(match_row, match_type, match_rules):
    try:
        c = conn.cursor()
        c.execute('''INSERT INTO matches 
                    (donation_id, mutation_id, donation_amount, unique_code, credit,
                     donor_name, bank_description, match_type, match_rules)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)''',
                 (str(match_row['donation_id']), 
                  str(match_row['mutation_id']),
                  int(match_row['donation_amount']),
                  int(match_row['unique_code']),
                  float(match_row['credit']),
                  str(match_row['full_name']),
                  str(match_row['description']),
                  match_type,
                  json.dumps(match_rules)))
        conn.commit()
        return c.lastrowid
    except Exception as e:
        logger.error(f"Error storing match: {e}")
        return None

def get_transaction_summary(matched_data, original_df1, transaction_type):
    """
    Generate a summary of transaction matches and statistics
    """
    amount_col = 'donation_amount' if transaction_type == "Donation" else 'topup_amount'
    
    # Calculate statistics
    total_amount = matched_data[amount_col].sum()
    avg_amount = matched_data[amount_col].mean()
    total_original = len(original_df1)
    total_matches = len(matched_data)
    match_rate = (total_matches / total_original * 100) if total_original > 0 else 0
    
    summary = f"""### Transaction Summary
- **Match Statistics**:
  - Total Source Transactions: {total_original:,}
  - Successfully Matched: {total_matches:,}
  - Match Rate: {match_rate:.1f}%

- **Amount Statistics**:
  - Total Amount: Rp {total_amount:,.0f}
  - Average Amount: Rp {avg_amount:,.0f}
  - Number of Transactions: {total_matches:,}"""
    
    return summary

# **🔹 Streamlit UI**
st.title("📂 CSV Matching Tool with Logs")

# **🔹 Transaction Type Selection**
transaction_type = st.selectbox("🔄 Select Transaction Type", ["Donation", "Top-Up"])

# **🔹 File Upload**
uploaded_file_1 = st.file_uploader("📤 Upload CSV File 1", type=["csv"])
uploaded_file_2 = st.file_uploader("📤 Upload CSV File 2", type=["csv"])

available_rules = ["Name Match", "Amount Match", "Time Match", "Permutation Match"]
selected_rules = st.multiselect("✅ Select Matching Rules (Order Matters)", available_rules, default=[])

matched_data = None

if uploaded_file_1 and uploaded_file_2 and selected_rules:
    logger.info("Processing CSV files...")

    try:
        df1 = pd.read_csv(uploaded_file_1)
        df2 = pd.read_csv(uploaded_file_2)

        # Determine if we should use permutation matching
        is_permutation = "Permutation Match" in selected_rules
        
        df1, df2 = preprocess_data(df1, df2, is_permutation, transaction_type)
        
        # Filter out Permutation Match from matching rules since it's handled in preprocessing
        matching_rules = [rule for rule in selected_rules if rule != "Permutation Match"]
        matched_data = perform_matching(df1, df2, matching_rules, transaction_type)

        if matched_data is not None and not matched_data.empty:
            st.write("✅ **Matching Records Found:**")
            
            # Format currency values in the dataframe
            display_data = matched_data.copy()
            for col in ['donation_amount', 'topup_amount', 'credit']:
                if col in display_data.columns:
                    display_data[col] = display_data[col].apply(lambda x: f"Rp {x:,}")
            
            # Show transaction summary
            summary = get_transaction_summary(matched_data, df1, transaction_type)
            st.markdown(summary)
            
            st.write("### Matched Transactions")
            st.dataframe(display_data)
        else:
            st.warning("⚠️ No matching records found.")

    except Exception as e:
        error_msg = f"❌ Error: {e}"
        logger.error(error_msg, exc_info=True)
        st.error(error_msg)

st.sidebar.markdown("---")
if st.sidebar.button("View Historical Matches"):
    c = conn.cursor()
    c.execute('''SELECT * FROM matches ORDER BY created_at DESC LIMIT 10''')
    historical_matches = c.fetchall()
    
    if historical_matches:
        st.write("### Recent Matches")
        for match in historical_matches:
            st.write(f"**Match ID: {match[0]}**")
            st.write(f"Donation Amount: {match[3]}, Unique Code: {match[4]}, Credit: {match[5]}")
            st.write(f"Donor: {match[6]}")
            st.write(f"Bank Description: {match[7]}")
            st.write(f"Match Type: {match[8]}")
            st.write(f"Rules Applied: {match[9]}")
            st.write("---")
    else:
        st.write("No historical matches found")

conn.close()
