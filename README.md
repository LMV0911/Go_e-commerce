<div align="center">
  <img src="https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_Blue.png" alt="Golang Logo" width="150" />
  <h1>🛒 Projet E-Commerce CLI</h1>
  <p>Une plateforme de vente en ligne accessible directement depuis votre terminal.</p>
</div>

---

## 📖 À propos du projet

Ce projet a pour but de créer une application e-commerce fonctionnant entièrement en ligne de commande. 
Il vise à simuler les fonctionnalités classiques d'une boutique en ligne (navigation, panier, authentification, gestion des commandes) tout en conservant la légèreté et la rapidité du terminal.

## ✨ Fonctionnalités globales

L'application est divisée en plusieurs grands espaces :

- **Espace Client** : 
  - Création de compte et authentification.
  - Recherche et consultation du catalogue de produits.
  - Gestion du panier d'achat.
  - Validation de commande et simulation de paiement.
- **Espace Administration** : 
  - Gestion des stocks et des produits du catalogue.
  - Suivi et mise à jour des commandes clients.
  - Gestion des utilisateurs.
- **Serveur central** : 
  - API HTTP pour synchroniser et stocker les données.

## 🛠️ Installation et Lancement

1. **Cloner le projet**
   ```bash
   git clone https://github.com/LMV0911/Go_e-commerce.git
   cd Go_e-commerce
   ```

2. **Lancer le Serveur**
   ```bash
   go run ./cmd/server
   ```

3. **Lancer le Client (dans un autre terminal)**
   ```bash
   go run ./cmd/client
   ```
