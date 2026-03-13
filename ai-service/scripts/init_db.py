#!/usr/bin/env python3
"""Database initialization script for FlyMind AI Service.

This script sets up the database schema for the AI service.
"""

import asyncio
import logging
import sys
import os

# Add src to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src'))

from config import settings


logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)


async def init_database():
    """Initialize the database for the AI service."""
    logger.info("Initializing database...")

    # For Phase 1, we don't have specific tables for the AI service
    # The service uses Redis for caching and will use PostgreSQL
    # for future vector storage

    logger.info("Database initialization complete.")
    logger.info("Note: Vector storage tables will be created in Phase 2.")


async def main():
    """Main entry point."""
    try:
        await init_database()
    except Exception as e:
        logger.error(f"Failed to initialize database: {e}")
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
