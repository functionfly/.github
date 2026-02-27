import { Injectable, OnModuleInit } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { getDatabase, Database } from '../db';

@Injectable()
export class DatabaseService implements OnModuleInit {
  private db: Database;

  constructor(private configService: ConfigService) {}

  onModuleInit() {
    this.db = getDatabase(this.configService);
  }

  getDatabase(): Database {
    return this.db;
  }
}