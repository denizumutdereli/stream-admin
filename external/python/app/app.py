"""
TODO: dynamic queries of sources with parameters on demand over golang and saving training models to external db for reuse.
TODO: continues learning intervals and enhancements with compare benchmarks
TODO: Threshold and response with anomalous predictions 
Current Score Examinations are:

-50 recent orders excluding user_id 1

Mean Score: Approximately 0.068
Median Score: Approximately 0.105
Minimum Score: -0.167 (most anomalous)
Maximum Score: 0.158 (least anomalous)
Standard Deviation: Approximately 0.082

"""

import os

import pandas as pd
from flask import Flask, jsonify, request
from flask_sqlalchemy import SQLAlchemy
from joblib import dump, load
from sklearn.ensemble import IsolationForest
from sklearn.preprocessing import StandardScaler

app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = 'postgresql://citus:citus@host.minikube.internal:5433/analytics'
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
db = SQLAlchemy(app)

MODEL_PATH = 'model.joblib'
SCALER_PATH = 'scaler.joblib'

@app.route('/train', methods=['POST'])
def train_model():
    query = "SELECT CASE WHEN side = 'BUY' THEN 0 ELSE 1 END as side, price, stop_price, quantity, " \
            "quote_asset_quantity, executed_quantity, cumulative_quote_quantity, " \
            "CASE WHEN status = 'NEW' THEN 0 WHEN status = 'PARTIALLY_FILLED' THEN 1 WHEN status = 'FILLED' THEN 2 " \
            "WHEN status = 'TIMED_OUT' THEN 3 ELSE 4 END as status, " \
            "0 as match_engine, commission, " \
            "CASE WHEN type = 'MARKET_LIMIT' THEN 0 WHEN type = 'STOP_LIMIT' THEN 1 WHEN type = 'LIMIT' THEN 2 " \
            "WHEN type = 'BUY' THEN 3 ELSE 4 END as type, commission_try, commission_usdt " \
            "FROM order_orders WHERE user_id != 1"
    df = pd.read_sql(query, db.engine)

 
    df = df.fillna(0) 
    scaler = StandardScaler()
    df_scaled = pd.DataFrame(scaler.fit_transform(df), columns=df.columns)

    model = IsolationForest(n_estimators=100, contamination='auto', random_state=42)
    model.fit(df_scaled)

    dump(model, MODEL_PATH)
    dump(scaler, SCALER_PATH)

    return jsonify({"message": "Model trained and saved successfully"})

@app.route('/analyze', methods=['POST'])
def analyze_data():
    if not os.path.exists(MODEL_PATH) or not os.path.exists(SCALER_PATH):
        return jsonify({"error": "Model or scaler file not found"}), 400

    model = load(MODEL_PATH)
    scaler = load(SCALER_PATH)

    data = request.json['data']
    df = pd.DataFrame(data)

    df['side'] = df['side'].map({'BUY': 0, 'SELL': 1})
    df['status'] = df['status'].map({'NEW': 0, 'PARTIALLY_FILLED': 1, 'FILLED': 2, 'TIMED_OUT': 3, 'CANCELLED': 4})
    df['match_engine'] = 0  # Since only 'EXTERNAL' is present
    df['type'] = df['type'].map({'MARKET_LIMIT': 0, 'STOP_LIMIT': 1, 'LIMIT': 2, 'BUY': 3, 'SELL': 4})
    df = df.fillna(0)
    
    relevant_fields = ['side', 'price', 'stop_price', 'quantity', 'quote_asset_quantity', 
                       'executed_quantity', 'cumulative_quote_quantity', 'status', 'match_engine', 
                       'commission', 'type', 'commission_try', 'commission_usdt']
    df_scaled = pd.DataFrame(scaler.transform(df[relevant_fields]), columns=relevant_fields)

    scores = model.decision_function(df_scaled)
    results = pd.DataFrame({'client_order_id': df['client_order_id'], 'anomaly_score': scores})
    return jsonify({"algorithm": "Isolation Forest", "results": results.to_dict(orient='records')})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=9090)
