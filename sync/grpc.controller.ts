import { Controller, Inject, Logger } from '@nestjs/common';
import { GrpcMethod } from '@nestjs/microservices';
import { Wallet, UserStats } from '../database/drizzle/schema';
import { NodePgDatabase } from 'drizzle-orm/node-postgres';
import * as schema from '../database/drizzle/schema';
import { eq } from 'drizzle-orm';
import { SyncResponse } from './grpc.service';

@Controller()
export class GRPCController {
  private readonly logger = new Logger(GRPCController.name);

  constructor(
    @Inject('DATABASE_CONNECTION')
    private db: NodePgDatabase<typeof schema>,
  ) {}

  /**
   * @param wallets - Array of wallet data to be synchronized.
   * @returns A SyncResponse indicating the result of the synchronization.
   *
   */
  @GrpcMethod('UserService', 'SyncWallets')
  async syncWallets(wallets: Partial<Wallet>[]): Promise<SyncResponse> {
    try {
      this.logger.log(`Received ${wallets.length} wallets for synchronization`);

      await this.db.transaction(async (tx) => {
        for (const wallet of wallets) {
          const { walletId, balance, createdBy, updatedBy } = wallet;

          if (!walletId) {
            throw new Error('Wallet ID is required for synchronization');
          }

          await tx
            .update(schema.wallets)
            .set({
              balance: balance,
              createdBy: createdBy,
              updatedBy: updatedBy,
              updatedAt: new Date(),
            })
            .where(eq(schema.wallets.walletId, walletId));
        }
      });

      return { success: true, message: 'Wallets synchronized successfully' };
    } catch (error) {
      this.logger.error(
        `Failed to sync wallets: ${error.message}`,
        error.stack,
      );
      return {
        success: false,
        message: `Failed to sync wallets: ${error.message}`,
      };
    }
  }

  /**
   *
   * @param stats - User statistics data to be synchronized.
   * @returns
   */
  @GrpcMethod('UserService', 'SyncUserStats')
  async syncUserStats(stats: Partial<UserStats>): Promise<SyncResponse> {
    try {
      this.logger.log(`Received  user stats for synchronization`);

      const [updatedStats] = await this.db.transaction(async (tx) => {
        const {
          userStatsId,
          createdBy,
          updatedBy,
          userStatsIdSerial,
          ...updateableFields
        } = stats;

        if (!userStatsId) {
          throw new Error('User Stats ID is required for synchronization');
        }

        // Remove any undefined values from updateableFields
        Object.keys(updateableFields).forEach(
          (key) =>
            updateableFields[key] === undefined && delete updateableFields[key],
        );

        return await tx
          .update(schema.usersStats)
          .set({
            ...updateableFields,
            updatedBy,
            updatedAt: new Date(),
          })
          .where(eq(schema.usersStats.userStatsId, userStatsId))
          .returning();
      });

      return {
        success: true,
        message: 'User stats synchronized successfully',
        data: updatedStats,
      };
    } catch (error) {
      this.logger.error(
        `Failed to sync user stats: ${error.message}`,
        error.stack,
      );
      return {
        success: false,
        message: `Failed to sync user stats: ${error.message}`,
      };
    }
  }
}
