<div align="center">
  <img src="https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_Blue.png" alt="Golang Logo" width="150" />
  <h1>🛒 Projet E-Commerce CLI</h1>
  <p>Une plateforme de vente en ligne accessible directement depuis votre terminal.</p>
</div>

---

## 📖 À propos du projet

Ce projet a pour but de créer une application e-commerce fonctionnant entièrement en ligne de commande. 
Il simule les fonctionnalités d'une boutique en ligne tout en conservant la légèreté et la rapidité du terminal.

## ✨ Fonctionnalités globales

L'application est divisée en plusieurs espaces :

- **Espace Client** : 
  - Création de compte et authentification.
  - Recherche et consultation du catalogue de produits.
  - Gestion du panier d'achat (BSK-XXXXXX).
  - Validation de commande (CMD-XXXXXX) et simulation de paiement.
- **Espace Administration** : 
  - Gestion des produits et stocks avec identifiant unique (PDT-XXXXXX).
  - Suivi et mise à jour des commandes clients avec motif d'annulation obligatoire.
  - Création manuelle de commande client par l'administrateur.
  - Gestion et confirmation manuelle des utilisateurs.
- **Serveur central** : 
  - API HTTP sans framework (module standard `net/http`).
  - Base de données PostgreSQL via Docker (`database/sql` natif sans ORM).
- **Accès Distant SSH (Wish)** :
  - Exposition distante de l'interface CLI sur le port 2222.

## 🛠️ Installation et Démarrage

1. **Cloner le dépôt**
   ```bash
   git clone https://github.com/LMV0911/Go_e-commerce.git
   cd Go_e-commerce
   ```

2. **Configurer l'environnement**
   Copiez le fichier d'exemple et renseignez vos variables sensibles (`POSTGRES_PASSWORD`, `JWT_SECRET`) :
   ```bash
   cp .env.example .env
   ```

3. **Démarrer les services (PostgreSQL, Adminer, Serveur HTTP & SSH)**
   ```bash
   docker compose up -d
   ```

4. **Lancer la CLI Admin (Bubbletea)**
   ```bash
   docker compose run --rm admin /app/bin/admin
   ```

5. **Accéder à la CLI via SSH (Wish)**
   ```bash
   ssh -p 2222 admin@localhost
   ```

## 📡 Endpoints API Admin (`net/http`)

| Méthode | Route | Description |
| :--- | :--- | :--- |
| `GET` | `/health` | Vérification de l'état du serveur |
| `GET` | `/api/admin/products` | Liste de tous les produits du catalogue |
| `GET` | `/api/admin/products/{id}` | Détail d'un produit par identifiant (`PDT-XXXXXX`) |
| `POST` | `/api/admin/products` | Création d'un produit (génération automatique de l'ID) |
| `PUT` | `/api/admin/products/{id}` | Modification d'un produit |
| `DELETE` | `/api/admin/products/{id}` | Suppression d'un produit |
| `GET` | `/api/admin/users` | Liste des utilisateurs enregistrés |
| `GET` | `/api/admin/users/{id}` | Détail d'un compte utilisateur |
| `POST` | `/api/admin/users/{id}/confirm` | Validation / confirmation manuelle d'un compte |
| `PUT` | `/api/admin/users/{id}` | Mise à jour des informations utilisateur |
| `DELETE` | `/api/admin/users/{id}` | Suppression d'un compte utilisateur |
| `GET` | `/api/admin/orders` | Liste des commandes clients |
| `GET` | `/api/admin/orders/{id}` | Détail d'une commande avec ses articles |
| `PUT` | `/api/admin/orders/{id}/status` | Mise à jour du statut (motif obligatoire si annulé) |
| `POST` | `/api/admin/orders/manual` | Création manuelle d'une commande pour un client |
