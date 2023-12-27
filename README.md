# Stream-Admin: Comprehensive Data Management System

## Overview

Stream-Admin is an advanced data and users administrator system that integrates the functionalities of MPPS and the service infrastructure of a real-time data processing pipeline and overal. This platform is tailored for high-performance data analytics and real-time data distribution, primarily focusing on cryptocurrency markets. It leverages Citus for horizontal scalability of PostgreSQL, ensuring efficient data warehousing and streamlined data processing.

## Key Features

- **Citus-Powered Data Warehousing**: Utilizes Citus to scale PostgreSQL databases horizontally, facilitating efficient management of large-scale data.
- **Real-Time Data Processing**: Incorporates a robust service infrastructure built with Golang, capable of handling intricate data processing and distribution tasks in real-time.
- **CDC and Kafka Integration**: Seamlessly integrates CDC (Change Data Capture) connectors with Kafka streams for real-time data analytics and processing.
- **Advanced Data Querying and Management**: Employs domain-specific languages (DSL) for enhanced data querying, supporting complex data management operations.
- **Anomaly Detection with Python**: Implements Python-based anomaly detection using the Isolation Forest algorithm, essential for maintaining data integrity and security.
- **DSL Powered Filters and Data Processing**: Implements DSL pegyjs and Langchain for NLP processing with direct user interaction and query on behalf.
- **Scalable and Resilient Architecture**: Designed for high availability and scalability, using technologies like Kubernetes for orchestration and Nats for distributed messaging.

## System Architecture

Stream-Admin's architecture synergizes Citus's distributed data processing capabilities with a Golang-based service infrastructure. It ensures efficient handling of both analytical and transactional data workloads, supported by a scalable and resilient framework.

### Core Components

- **Citus for PostgreSQL**: Horizontally scales out PostgreSQL for handling massive data workloads.
- **Real-Time Processing Pipeline**: A Golang-based infrastructure for processing and distributing cryptocurrency market data.
- **CDC & Kafka**: For capturing and streaming data changes in real-time, facilitating immediate data analytics.
- **Python-based Anomaly Detection**: Utilizes Python for analyzing and scoring data anomalies with using Kubernetes API and Isolation Forest.
- **Langchain**: LangChain is a framework designed to simplify the creation of applications using large language models. TBC

## MPP -Citus in details

MPP harnesses the strength of a central MPP (Massively Parallel Processing) database, merging disparate microservice databases into a single, powerful data warehouse. This system capitalizes on Citus's ability to horizontally scale PostgreSQL across multiple machines, transforming it into a distributed database, a high-performance analytical database (OLAP), and a formidable transactional system (OLTP) when necessary.

Long story short, all microservice databases in one performant place as tables.

By integrating CDC (Change Data Capture) connectors, Kafka streams, and JDBC sinks, mpp-citus-pulse provides real-time data integration and analytics processing, significantly enhancing query performance and system reliability. Full isolated from production databases and synced.

## MPP Key Concept ##

Our architecture is built on the foundational concept of data sharding. Citus intelligently partitions your database across several nodes, distributing the load and enabling parallelized query execution. This approach not only increases performance—often by more than 300x compared to standard PostgreSQL—but also enhances the system's fault tolerance and scalability.

Moreover, we've isolated analytical processes from transactional ones, ensuring operational databases (OLTP) are not bogged down by heavy analytical queries. By using CDC and Kafka, changes in the operational databases are captured and streamed in real-time to Citus, ensuring the analytical system (OLAP) can leverage up-to-the-moment data without impacting operational performance.


## MPP Features
- **CDC Connectors**: Capture data changes in real-time using Debezium.
- **Kafka**: Ensure reliable and scalable data streaming.
- **KSQL**: KSQL is a streaming SQL engine for Apache Kafka. An interactive SQL interface for processing data in Kafka.
- **JDBC Sink**: Stream data into your database with the Kafka JDBC Sink Connector.
- **Citus MPP**: Scale out your PostgreSQL database horizontally with Citus.

## MPP Components
- [Citus MPP](https://www.citusdata.com/)
- [Debezium CDC](https://debezium.io/)
- [Avro Schemas](https://avro.apache.org/)
- [Kafka Schema Registry](https://docs.confluent.io/platform/current/schema-registry/index.html)
- [Kafka Connectors](https://docs.confluent.io/platform/current/connect/kafka_connectors.html)
- [Kafka JDBC Sink Connector](https://docs.confluent.io/kafka-connectors/jdbc/current/sink-connector/overview.html)
- [KSQL Streams](https://www.confluent.io/blog/ksql-streaming-sql-for-apache-kafka/)


## Getting Started

### Prerequisites

- Docker and Docker Compose
- Kubernetes Cluster
- Kafka and Nats setup

### Installation

\```bash
docker-compose up -d
# Follow-up commands and configurations...
\```

## Usage

### Basic Operations

- Initialize the system: `stream-admin init`
- Start data processing: `stream-admin start`
- Monitor system performance: `stream-admin monitor`

## Data Management

Leverages the combined power of Citus and Kafka for managing large-scale data. It ensures real-time data synchronization across different microservice databases, aligning them into a cohesive data warehouse.

## Security and Compliance

Stream-Admin adheres to strict security protocols, incorporating RBAC (Role-Based Access Control) for secure data operations and compliance with data privacy regulations.

## Monitoring and Logging

Employs Prometheus for system monitoring and the EFK stack for logging, ensuring comprehensive oversight of system health and performance.

## External Resources

- [Citus Documentation](https://www.citusdata.com/docs)
- [Kafka Documentation](https://kafka.apache.org/documentation/)
- [Kubernetes Documentation](https://kubernetes.io/docs/home/)

## Contributing

Contributions are welcome. Please follow the contributing guidelines to participate.

## License

[@denizumutdereli](https://www.linkedin.com/in/denizumutdereli)