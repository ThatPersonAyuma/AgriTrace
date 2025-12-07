from flask import Flask, request, jsonify, render_template
from flask_cors import CORS

app = Flask(__name__)
CORS(app)  # mengizinkan semua origin

# Simulasi penyimpanan job
jobs = {}

# Route untuk membuka HTML
@app.route('/')
def index():
    return render_template('delivery.html')  # folder templates/delivery.html
@app.route('/check')
def check():
    return render_template('delivery_check.html')
@app.route('/test')
def test():
    return render_template('test.html')
@app.route('/order')
def order():
    return render_template('order.html')
@app.route('/product')
def product():
    return render_template('product.html')
@app.route('/check/checkpoint')
def checkpoint():
    return render_template('comfirm_checkpoint.html')
# API endpoint untuk create shipment
# @app.route('/logistic/create', methods=['POST'])
# def create_shipment():
#     data = request.json
#     job_id = str(uuid.uuid4())

#     if not data:
#         jobs[job_id] = {"status": "error", "error": "Invalid request"}
#         return jsonify({"job_id": job_id, "status": "error", "error": "Invalid request"}), 400

#     jobs[job_id] = {"status": "submitted", "payload": data}
    
#     return jsonify({"job_id": job_id, "status": "submitted"})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8020, debug=True)
