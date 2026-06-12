# HỆ THỐNG QUẢN LÝ PHÒNG KHÁM / BỆNH VIỆN (HIS + LIS + RIS + PACS + EMR)

## 1. Mục tiêu dự án

Xây dựng nền tảng quản lý bệnh viện/phòng khám theo chuẩn HIS hiện đại.

### Công nghệ

Backend:

* Golang

Database:

* PostgreSQL
* MongoDB
* Redis

Message Queue:

* Redis Stream

Desktop:

* Wails
* React
* Typescript

Triển khai:

* Docker
* Kubernetes

---

# 2. Kiến trúc tổng thể

## Backend

API Gateway

Các Domain Service:

* Identity & Access
* Patient Management
* Appointment
* Reception
* EMR
* LIS
* RIS
* PACS
* Pharmacy
* Billing
* Inventory
* Inpatient
* Reporting
* Notification
* Audit

## Database

PostgreSQL

Lưu dữ liệu giao dịch:

* User
* Patient
* Appointment
* Visit
* Billing
* Inventory

MongoDB

Lưu dữ liệu phi cấu trúc:

* Medical Record
* Lab Result
* Radiology Result
* DICOM Metadata
* Audit Log
* Documents

Redis

* Session
* OTP
* Cache
* Queue

---

# 3. HỆ THỐNG HIS

## 3.1 Identity & Access

Chức năng

* Đăng nhập
* Đăng xuất
* Refresh Token
* Quản lý User
* RBAC
* MFA

Bảng dữ liệu

users
roles
permissions
user_roles
role_permissions

API

POST /auth/login

POST /auth/logout

POST /auth/refresh

GET /users

POST /users

PUT /users/{id}

DELETE /users/{id}

---

## 3.2 Patient Management

Chức năng

* Hồ sơ bệnh nhân
* Tiền sử bệnh
* Người thân
* BHYT
* CCCD

Bảng dữ liệu

patients
patient_contacts
patient_insurance
patient_history

API

POST /patients

GET /patients

GET /patients/{id}

PUT /patients/{id}

---

## 3.3 Appointment

Chức năng

* Đặt lịch
* Hủy lịch
* Đổi lịch

Bảng

appointments
appointment_slots

API

POST /appointments

PUT /appointments/{id}

DELETE /appointments/{id}

---

## 3.4 Reception

Chức năng

* Check-in
* Phát số thứ tự
* Chuyển phòng khám

Bảng

queues
queue_calls

API

POST /checkin

POST /queue/call

---

## 3.5 Outpatient Visit

Chức năng

* Khám bệnh
* Chẩn đoán
* Chỉ định

Bảng

visits
diagnoses
visit_notes

API

POST /visits

POST /visits/{id}/diagnosis

POST /visits/{id}/order

---

# 4. EMR

## Hồ sơ bệnh án điện tử

Dữ liệu MongoDB

medical_records

Mẫu document

* SOAP
* Vitals
* Diagnosis
* Orders
* Prescription
* Attachments

Versioning

Mỗi lần sửa bệnh án:

* version +1

Lưu toàn bộ lịch sử

Audit

* ai sửa
* sửa gì
* thời gian

---

# 5. LIS

Laboratory Information System

## Chức năng

* Tạo phiếu xét nghiệm
* Nhận mẫu
* Chạy mẫu
* Trả kết quả

Workflow

Bác sĩ chỉ định

↓

Phiếu xét nghiệm

↓

Lấy mẫu

↓

Máy xét nghiệm

↓

Kết quả

↓

Bác sĩ xem

Bảng

lab_orders
lab_samples
lab_results

MongoDB

lab_result_documents

API

POST /labs/orders

POST /labs/samples

POST /labs/results

GET /labs/results/{id}

---

# 6. RIS

Radiology Information System

## Chức năng

* Chỉ định chẩn đoán hình ảnh
* Quản lý ca chụp
* Trả kết quả

Bảng

radiology_orders
radiology_reports

MongoDB

radiology_documents

API

POST /radiology/orders

POST /radiology/report

GET /radiology/report/{id}

---

# 7. PACS

Picture Archiving and Communication System

## Chức năng

* Lưu ảnh DICOM
* Truy xuất ảnh
* Viewer

Storage

Object Storage

* MinIO
* S3

Metadata

MongoDB

dicom_metadata

API

POST /pacs/upload

GET /pacs/study/{id}

GET /pacs/series/{id}

GET /pacs/image/{id}

---

# 8. Pharmacy

## Chức năng

* Kê đơn
* Xuất thuốc
* Tồn kho

Bảng

drug_catalog
prescriptions
prescription_items

API

POST /prescriptions

POST /pharmacy/dispense

GET /prescriptions/{id}

---

# 9. Inventory

## Chức năng

* Kho thuốc
* Kho vật tư

Bảng

warehouses
inventory_items
stock_transactions

API

POST /inventory/import

POST /inventory/export

GET /inventory/stock

---

# 10. Billing

## Chức năng

* Thu viện phí
* Thu thuốc
* BHYT

Bảng

invoices
invoice_items
payments

API

POST /invoice

POST /payment

GET /invoice/{id}

---

# 11. Inpatient

## Chức năng

* Quản lý giường
* Nhập viện
* Xuất viện

Bảng

wards
rooms
beds
admissions

API

POST /admission

POST /discharge

GET /beds

---

# 12. Reporting

## Dashboard

* Doanh thu
* Số bệnh nhân
* Tồn kho
* Công suất giường

MongoDB

report_snapshots

---

# 13. Notification

## Chức năng

* SMS
* Email
* Push Notification

Queue

Redis Stream

Topic

appointment.created

lab.completed

radiology.completed

invoice.created

---

# 14. Audit

## Chức năng

Theo dõi toàn bộ hệ thống

MongoDB

audit_logs

Thông tin

* User
* Action
* Entity
* Timestamp
* IP
* Device

---

# 15. THIẾT KẾ DATABASE

PostgreSQL

Dự kiến 150-200 bảng

Nhóm bảng:

Identity
≈10 bảng

Patient
≈15 bảng

Appointment
≈10 bảng

Visit
≈20 bảng

EMR Metadata
≈10 bảng

LIS
≈15 bảng

RIS
≈10 bảng

PACS Metadata
≈10 bảng

Pharmacy
≈20 bảng

Inventory
≈20 bảng

Billing
≈15 bảng

Inpatient
≈20 bảng

Reporting
≈10 bảng

Audit
≈5 bảng

---

# 16. API GATEWAY

Gateway

/api/v1

Ví dụ

/api/v1/auth

/api/v1/patients

/api/v1/appointments

/api/v1/visits

/api/v1/labs

/api/v1/radiology

/api/v1/pacs

/api/v1/pharmacy

/api/v1/billing

Gateway Responsibilities

* Authentication
* Authorization
* Rate Limit
* Audit
* Request Logging
* Response Logging
* Tracing

---

# 17. DDD ARCHITECTURE

hospital-system

cmd/

api

worker

migration

internal/

identity/

patient/

appointment/

reception/

visit/

emr/

lis/

ris/

pacs/

pharmacy/

inventory/

billing/

inpatient/

reporting/

notification/

audit/

Mỗi module

domain/

application/

infrastructure/

presentation/

---

Ví dụ

patient/

domain/

entities/

repositories/

services/

events/

application/

commands/

queries/

handlers/

dto/

infrastructure/

postgres/

mongodb/

cache/

presentation/

http/

grpc/

---

# 18. EVENT DRIVEN ARCHITECTURE

Domain Events

PatientCreated

AppointmentCreated

VisitCreated

LabOrderCreated

LabResultCompleted

RadiologyCompleted

PrescriptionCreated

InvoiceCreated

Event Bus

Redis Stream

---

# 19. BẢO MẬT

JWT

Access Token

Refresh Token

RBAC

MFA

AES-256

Audit Log

Backup

Disaster Recovery

Encryption At Rest

Encryption In Transit

TLS 1.3

---

# 20. ROADMAP

PHASE 1

Identity

Patient

Appointment

Reception

Visit

EMR

PHASE 2

Pharmacy

Inventory

Billing

Reporting

PHASE 3

LIS

RIS

PACS

PHASE 4

Inpatient

Telehealth

Mobile App

PHASE 5

EMR chuẩn Bộ Y tế

Chữ ký số

BHYT

HL7

FHIR

DICOM
