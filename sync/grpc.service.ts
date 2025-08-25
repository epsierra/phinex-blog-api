import { Inject, Injectable, OnModuleInit } from '@nestjs/common';
import { ClientGrpc } from '@nestjs/microservices';
import { Observable, lastValueFrom } from 'rxjs';
import {
  Role,
  User,
  UserRole,
  UserStats,
  Wallet,
} from '../database/drizzle/schema';

interface UserSyncService {
  // User operations
  syncUser(data: any): Observable<any>;
  syncUserStats(data: any): Observable<any>;
  deleteUser(data: { user_id: string }): Observable<any>;
  updateUser(data: { user_id: string; user: any }): Observable<any>;

  // Role operations
  syncRole(data: any): Observable<any>;
  deleteRole(data: { role_id: string }): Observable<any>;
  updateRole(data: { role_id: string; role: any }): Observable<any>;

  // UserRole operations
  syncUserRole(data: any): Observable<any>;
  deleteUserRole(data: { user_role_id: string }): Observable<any>;
  updateUserRole(data: {
    user_role_id: string;
    user_role: any;
  }): Observable<any>;

  // Wallet operations
  syncWallet(data: any): Observable<any>;
  syncWallets(data: any): Observable<any>;
}

export type SyncResponse = {
  message: string;
  success: boolean;
  data?: Record<string, any>;
};

@Injectable()
export class GrpcUserSyncService implements OnModuleInit {
  private userService: UserSyncService;

  constructor(@Inject('USER_PACKAGE') private client: ClientGrpc) {}

  onModuleInit() {
    this.userService = this.client.getService<UserSyncService>('UserSync');
  }

  async syncUser(user: User): Promise<SyncResponse> {
    return lastValueFrom(this.userService.syncUser(user));
  }

  async syncUserStats(userStats: UserStats): Promise<SyncResponse> {
    return lastValueFrom(this.userService.syncUserStats(userStats));
  }

  async deleteUser(userId: string): Promise<SyncResponse> {
    return lastValueFrom(this.userService.deleteUser({ user_id: userId }));
  }

  async updateUser(userId: string, user: User): Promise<SyncResponse> {
    return lastValueFrom(
      this.userService.updateUser({
        user_id: userId,
        user,
      }),
    );
  }

  // Role methods
  async syncRole(role: Role): Promise<SyncResponse> {
    return lastValueFrom(this.userService.syncRole(role));
  }

  async deleteRole(roleId: string): Promise<SyncResponse> {
    return lastValueFrom(this.userService.deleteRole({ role_id: roleId }));
  }

  async updateRole(roleId: string, role: Role): Promise<SyncResponse> {
    return lastValueFrom(
      this.userService.updateRole({
        role_id: roleId,
        role,
      }),
    );
  }

  // UserRole methods
  async syncUserRole(userRole: UserRole): Promise<any> {
    return lastValueFrom(this.userService.syncUserRole(userRole));
  }

  async deleteUserRole(userRoleId: string): Promise<any> {
    return lastValueFrom(
      this.userService.deleteUserRole({ user_role_id: userRoleId }),
    );
  }

  async updateUserRole(userRoleId: string, userRole: UserRole): Promise<any> {
    return lastValueFrom(
      this.userService.updateUserRole({
        user_role_id: userRoleId,
        user_role: userRole,
      }),
    );
  }

  // Wallet methods
  async syncWallet(wallet: Wallet): Promise<SyncResponse> {
    return lastValueFrom(this.userService.syncWallet(wallet));
  }

  async syncWallets(wallets: Wallet[]): Promise<SyncResponse> {
    return lastValueFrom(this.userService.syncWallets({ wallets }));
  }
}
