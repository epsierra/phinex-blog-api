import { Injectable, Inject, NotFoundException, BadRequestException } from '@nestjs/common';
import { NodePgDatabase } from 'drizzle-orm/node-postgres';
import * as schema from '../database/drizzle/schema';
import { eq, sql, inArray, ilike, and } from 'drizzle-orm';
import {
  UserUpdateDto,
  CreateUserDto,
  CreateUserSettingDto,
  UpdateUserSettingDto,
} from './dto/user.dto';
import { customAlphabet } from 'nanoid';
import { ICurrentUser, Role, Role as RoleName, UserStatus, SubscriptionPlan, SubscriptionStatus } from '../common/constants';
import { addMonths, addYears } from 'date-fns';
import { AuthService } from '../auth/auth.service';
import * as bcrypt from 'bcrypt';
import { CustomLogger } from '../config/logger.config';
import { fromZonedTime } from 'date-fns-tz';
import { DatabaseErrorService } from '../common/services/error.service';

// Custom ID generator for lowercase alphanumeric IDs with 'phi' prefix, total 25 characters
const generateId = customAlphabet('abcdefghijklmnopqrstuvwxyz0123456789', 22);

@Injectable()
export class UsersService {
  private readonly logger = new CustomLogger();

  constructor(
    @Inject('DATABASE_CONNECTION')
    private db: NodePgDatabase<typeof schema>,
    private authService: AuthService,
    private databaseErrorService: DatabaseErrorService,
  ) {}

  /**
   * Helper method to format paginated responses.
   * @param data - The data array.
   * @param totalItems - Total number of items.
   * @param page - Current page number.
   * @param limit - Items per page.
   * @returns Paginated response object.
   */
  private paginateResponse<T>(
    data: T[],
    totalItems: number,
    page: number,
    limit: number,
  ) {
    const totalPages = Math.ceil(totalItems / limit);

    return {
      data,
      metadata: {
        currentPage: page,
        itemsPerPage: limit,
        totalItems,
        totalPages,
        hasNextPage: page < totalPages,
        hasPreviousPage: page > 1,
      },
    };
  }

  /**
   * Removes password from user object for security.
   * @param user User object with optional password.
   * @returns User object without password.
   */
  private excludePassword<T extends { password: string }>(
    user: T,
  ): Omit<T, 'password'> {
    const { password, ...userWithoutPassword } = user;
    return userWithoutPassword;
  }

  /**
   * Retrieves all users with pagination.
   * @param page Page number.
   * @param limit Items per page.
   * @param currentUser User performing the action.
   * @returns Paginated user data.
   */
  async findAll(page: number, limit: number, currentUser: ICurrentUser) {
    try {
      const offset = (page - 1) * limit;

      return await this.db.transaction(async (tx) => {
        // Fetch users and count concurrently
        const [users, count] = await Promise.all([
          tx.query.users.findMany({
            limit,
            offset,
          }),
          tx.select({ count: sql<number>`count(*)` }).from(schema.users),
        ]);

        const totalItems = Number(count[0].count);
        return this.paginateResponse(
          users.map((user) => this.excludePassword(user)),
          totalItems,
          page,
          limit,
        );
      });
    } catch (error) {
      this.logger.error(`Error fetching users: ${error.message}`, error.stack);
      this.databaseErrorService.handleDrizzleError(error, 'users retrieval');
    }
  }

  /**
   * Retrieves a single user by ID with detailed information including personal, contact, posts, followers, and followings.
   * @param userId User ID.
   * @param currentUser User performing the action.
   * @param page Page number for posts pagination.
   * @param limit Items per page for posts.
   * @returns Detailed user data with followers, followings, posts, and likes counts.
   */
  async findOne(
    userId: string,
    currentUser: ICurrentUser,
    page: number = 1,
    limit: number = 10,
  ) {
    try {
      return await this.db.transaction(async (tx) => {
        // Fetch user with related data
        const user = await tx.query.users.findFirst({
          where: eq(schema.users.userId, userId),
          columns: { password: false },
          with: {
            userAuth: true,
            userSettings: true,
            followers: true,
            followings: true,
            userRoles: {
              with: {
                role: true,
              },
            },
          },
        });

        if (!user) {
          throw new NotFoundException(`User with ${userId} does not exist`);
        }

        // Fetch posts with pagination
        const offset = (page - 1) * limit;
        const [blogs, blogsCount] = await Promise.all([
          tx.query.blogs.findMany({
            where: eq(schema.blogs.userId, userId),
            limit,
            offset,
          }),
          tx
            .select({ count: sql<number>`count(*)` })
            .from(schema.blogs)
            .where(eq(schema.blogs.userId, userId)),
        ]);

        const blogIds = blogs.map((blog) => blog.blogId);
        const totalBlogs = Number(blogsCount[0].count);

        // Fetch likes count for user's blogs
        const totalLikesResult = blogIds.length
          ? await tx
              .select({ count: sql<number>`count(*)` })
              .from(schema.likes)
              .where(inArray(schema.likes.refId, blogIds))
          : [{ count: 0 }];

        // Check if current user follows the target user
        const following = await tx.query.follows.findFirst({
          where: sql`${schema.follows.followerId} = ${currentUser.userId} AND ${schema.follows.followingId} = ${userId}`,
        });

        return {
          user: {
            ...user,
            following: !!following,
          },
          followersCount: Number(user.followers.length),
          followingsCount: Number(user.followings.length),
          totalLikes: Number(totalLikesResult[0].count),
          totalBlogs,
        };
      });
    } catch (error) {
      this.logger.error(`Error finding user: ${error.message}`, error.stack);
      this.databaseErrorService.handleDrizzleError(error, 'user retrieval');
    }
  }

  /**
   * Creates a new user with associated data and assigns specified role.
   * @param dto User creation data.
   * @param currentUser User performing the action.
   * @returns Created user data.
   */
  async createUser(dto: CreateUserDto, currentUser: ICurrentUser) {
    try {
      return await this.db.transaction(async (tx) => {
        // Check for email and phone number conflicts
        await this.authService.checkUserConflicts({
          email: dto.email,
          phoneNumber: dto.phoneNumber,
        });

        const userId = `phi${generateId()}`;
        const userAuthId = `phi${generateId()}`;
        const userRoleId = `phi${generateId()}`;
        const now = fromZonedTime(new Date(), 'UTC');

        // Hash password for security
        const hashedPassword = await bcrypt.hash(dto.password, 10);

        // Find or create the specified role
        let role = await tx.query.roles.findFirst({
          where: eq(schema.roles.roleName, dto.roleName as RoleName),
        });

        if (!role) {
          const roleId = `phi${generateId()}`;
          [role] = await tx
            .insert(schema.roles)
            .values({
              roleId,
              roleName: Role.Authenticated,
              createdBy: currentUser.fullName,
              createdAt: now,
              updatedBy: currentUser.fullName,
              updatedAt: now,
            })
            .returning();
        }

        // Insert user record
        const [baseUser] = await tx
          .insert(schema.users)
          .values({
            userId,
            firstName: dto.firstName,
            lastName: dto.lastName,
            fullName: dto.fullName,
            middleName: dto.middleName,
            phoneNumber: dto.phoneNumber,
            email: dto.email,
            password: hashedPassword,
            status: dto.status || 'active',
            verified: dto.verified || true,
            emailIsVerified: dto.emailIsVerified || true,
            phoneNumberIsVerified: dto.phoneNumberIsVerified || false,
            gender: dto.gender,
            dob: dto.dob,
            createdBy: currentUser.fullName,
            createdAt: now,
            updatedBy: currentUser.fullName,
            updatedAt: now,
          })
          .returning();

        // Insert user auth record
        await tx.insert(schema.userAuths).values({
          userAuthId,
          userId,
          authProvider: 'email',
          createdBy: currentUser.fullName,
          createdAt: now,
          updatedBy: currentUser.fullName,
          updatedAt: now,
        });

        // Assign role to user
        await tx.insert(schema.userRoles).values({
          userRoleId,
          userId,
          roleId: role.roleId,
          createdBy: currentUser.fullName,
          createdAt: now,
          updatedBy: currentUser.fullName,
          updatedAt: now,
        });

        // Check if a wallet already exists for the user, if not, create one
        const existingWallet = await tx.query.wallets.findFirst({
          where: eq(schema.wallets.userId, userId),
        });

        if (!existingWallet) {
          const walletId = `phi${generateId()}`;
          await tx.insert(schema.wallets).values({
            walletId,
            userId,
            createdBy: currentUser.fullName,
            createdAt: now,
            updatedBy: currentUser.fullName,
            updatedAt: now,
          });
        }

        return {
          message: 'User created successfully',
          data: this.excludePassword(baseUser),
        };
      });
    } catch (error) {
      this.logger.error(`Error creating user: ${error.message}`, error.stack);
      this.databaseErrorService.handleDrizzleError(error, 'user creation');
    }
  }

  /**
   * Updates a user's information.
   * @param userId User ID.
   * @param updateDto Update data.
   * @param currentUser User performing the action.
   * @returns Updated user data.
   */
  async update(
    userId: string,
    updateDto: UserUpdateDto,
    currentUser: ICurrentUser,
  ) {
    try {
      return await this.db.transaction(async (tx) => {
        // Fetch existing user
        const existingUser = await tx.query.users.findFirst({
          where: eq(schema.users.userId, userId),
        });

        if (!existingUser) {
          throw new NotFoundException('User not found');
        }

        // Check for conflicts in updated fields
        await this.authService.checkUserConflicts({
          email: updateDto.email,
          phoneNumber: updateDto.phoneNumber,
          userId,
        });

        // Update user record
        const updatePayload: any = { ...updateDto };
        if (updateDto.password) {
          updatePayload.password = await bcrypt.hash(updateDto.password, 10);
        }
        updatePayload.modifiedBy = currentUser.fullName;
        updatePayload.updatedAt = fromZonedTime(new Date(), 'UTC');

        const [updatedUser] = await tx
          .update(schema.users)
          .set(updatePayload)
          .where(eq(schema.users.userId, userId))
          .returning();

        return {
          message: 'User updated successfully',
          data: this.excludePassword(updatedUser),
        };
      });
    } catch (error) {
      this.logger.error(`Error updating user: ${error.message}`, error.stack);
      this.databaseErrorService.handleDrizzleError(error, 'user update');
    }
  }

  /**
   * Deletes a user.
   * @param userId User ID.
   * @param currentUser User performing the action.
   * @returns Success message.
   */
  async remove(userId: string, currentUser: ICurrentUser) {
    try {
      return await this.db.transaction(async (tx) => {
        // Fetch existing user
        const existingUser = await tx.query.users.findFirst({
          where: eq(schema.users.userId, userId),
        });

        if (!existingUser) {
          throw new NotFoundException('User not found');
        }

        // Delete user
        await tx.delete(schema.users).where(eq(schema.users.userId, userId));

        return { message: 'User deleted successfully' };
      });
    } catch (error) {
      this.logger.error(`Error deleting user: ${error.message}`, error.stack);
      this.databaseErrorService.handleDrizzleError(error, 'user deletion');
    }
  }

  /**
   * Retrieves all roles assigned to a user.
   * @param userId User ID.
   * @returns Array of user roles.
   */
  async getUserRoles(userId: string) {
    try {
      // Fetch user roles with role details
      const userRoles = await this.db.query.userRoles.findMany({
        where: eq(schema.userRoles.userId, userId),
        with: {
          role: true,
        },
      });

      return userRoles.map((ur) => ur.role);
    } catch (error) {
      this.logger.error(
        `Error fetching user roles: ${error.message}`,
        error.stack,
      );
      this.databaseErrorService.handleDrizzleError(
        error,
        'user roles retrieval',
      );
    }
  }

  /**
   * Creates a new user setting.
   * @param userId User ID.
   * @param dto Setting data.
   * @param currentUser User performing the action.
   * @returns Created setting data.
   */
  async createUserSetting(
    userId: string,
    dto: CreateUserSettingDto,
    currentUser: ICurrentUser,
  ) {
    try {
      return await this.db.transaction(async (tx) => {
        // Insert user setting
        const userSettingId = `phi${generateId()}`;
        const [setting] = await tx
          .insert(schema.userSettings)
          .values({
            userSettingId,
            userId,
            pushNotificationsEnabled: dto.pushNotificationsEnabled ?? true,
            emailNotificationsEnabled: dto.emailNotificationsEnabled ?? true,
            theme: dto.theme ?? 'light',
            language: dto.language ?? 'en',
            profileVisibility: dto.profileVisibility ?? 'public',
            locationSharingEnabled: dto.locationSharingEnabled ?? false,
            autoRefreshEnabled: dto.autoRefreshEnabled ?? true,
            settings: dto.settings,
            createdBy: currentUser.fullName,
            createdAt: fromZonedTime(new Date(), 'UTC'),
            updatedBy: currentUser.fullName,
            updatedAt: fromZonedTime(new Date(), 'UTC'),
          })
          .returning();

        return {
          message: 'User setting created successfully',
          data: setting,
        };
      });
    } catch (error) {
      this.logger.error(
        `Error creating user setting: ${error.message}`,
        error.stack,
      );
      this.databaseErrorService.handleDrizzleError(
        error,
        'user setting creation',
      );
    }
  }

  /**
   * Retrieves all settings for a user.
   * @param userId User ID.
   * @returns Array of user settings.
   */
  async getUserSettings(userId: string) {
    try {
      // Fetch user settings
      const settings = await this.db.query.userSettings.findMany({
        where: eq(schema.userSettings.userId, userId),
      });

      return settings;
    } catch (error) {
      this.logger.error(
        `Error fetching user settings: ${error.message}`,
        error.stack,
      );
      this.databaseErrorService.handleDrizzleError(
        error,
        'user settings retrieval',
      );
    }
  }

  /**
   * Updates a user setting.
   * @param userId User ID.
   * @param userSettingId Setting ID.
   * @param dto Update data.
   * @param currentUser User performing the action.
   * @returns Updated setting data.
   */
  async updateUserSetting(
    userId: string,
    userSettingId: string,
    dto: UpdateUserSettingDto,
    currentUser: ICurrentUser,
  ) {
    try {
      return await this.db.transaction(async (tx) => {
        // Fetch existing setting
        const [oldSetting] = await tx
          .select()
          .from(schema.userSettings)
          .where(eq(schema.userSettings.userSettingId, userSettingId));

        if (!oldSetting) {
          throw new NotFoundException('Setting not found');
        }

        // Update setting
        const [setting] = await tx
          .update(schema.userSettings)
          .set({
            ...dto,
            updatedBy: currentUser.fullName,
            updatedAt: fromZonedTime(new Date(), 'UTC'),
          })
          .where(eq(schema.userSettings.userSettingId, userSettingId))
          .returning();

        return {
          message: 'User setting updated successfully',
          data: setting,
        };
      });
    } catch (error) {
      this.logger.error(
        `Error updating user setting: ${error.message}`,
        error.stack,
      );
      this.databaseErrorService.handleDrizzleError(
        error,
        'user setting update',
      );
    }
  }

  /**
   * Deletes a user setting.
   * @param userId User ID.
   * @param userSettingId Setting ID.
   * @param currentUser User performing the action.
   * @returns Success message.
   */
  async deleteUserSetting(
    userId: string,
    userSettingId: string,
    currentUser: ICurrentUser,
  ) {
    try {
      return await this.db.transaction(async (tx) => {
        // Fetch existing setting for audit logging
        const [existingSetting] = await tx
          .select()
          .from(schema.userSettings)
          .where(eq(schema.userSettings.userSettingId, userSettingId));

        if (!existingSetting) {
          throw new NotFoundException('Setting not found');
        }

        // Delete setting
        await tx
          .delete(schema.userSettings)
          .where(eq(schema.userSettings.userSettingId, userSettingId));

        return { message: 'User setting deleted successfully' };
      });
    } catch (error) {
      this.logger.error(
        `Error deleting user setting: ${error.message}`,
        error.stack,
      );
      this.databaseErrorService.handleDrizzleError(
        error,
        'user setting deletion',
      );
    }
  }

  /**
   * Updates a user's status.
   * @param userId User ID.
   * @param status The new status for the user.
   * @param currentUser User performing the action.
   * @returns Updated user data.
   */
  async updateUserStatus(
    userId: string,
    status: UserStatus,
    currentUser: ICurrentUser,
  ) {
    try {
      return await this.db.transaction(async (tx) => {
        const existingUser = await tx.query.users.findFirst({
          where: eq(schema.users.userId, userId),
        });

        if (!existingUser) {
          throw new NotFoundException('User not found');
        }

        const [updatedUser] = await tx
          .update(schema.users)
          .set({
            status: status,
            updatedBy: currentUser.fullName,
            updatedAt: fromZonedTime(new Date(), 'UTC'),
          })
          .where(eq(schema.users.userId, userId))
          .returning();

        return {
          message: 'User status updated successfully',
          data: this.excludePassword(updatedUser),
        };
      });
    } catch (error) {
      this.logger.error(
        `Error updating user status: ${error.message}`,
        error.stack,
      );
      this.databaseErrorService.handleDrizzleError(error, 'user status update');
    }
  }

  /**
   * Subscribes a user to a chosen plan.
   * @param userId - The ID of the user subscribing.
   * @param plan - The subscription plan (monthly, yearly, two_months_discount).
   * @param businessId - Optional: The ID of the business associated with the subscription.
   * @param currentUser - The current authenticated user's information.
   * @returns A success message upon successful subscription.
   * @throws {NotFoundException} If the user or their wallet is not found.
   * @throws {BadRequestException} If the wallet balance is insufficient.
   * @throws {InternalServerErrorException} If a database error occurs.
   */
  async subscribeToPlan(
    userId: string,
    plan: SubscriptionPlan,
    businessId: string | undefined,
    currentUser: ICurrentUser,
  ) {
    try {
      return await this.db.transaction(async (tx) => {
        const user = await tx.query.users.findFirst({
          where: eq(schema.users.userId, userId),
          with:{
            business:true
          }
        });

        if (!user) {
          throw new NotFoundException(`User with ID ${userId} not found.`);
        }

        const userWallet = await tx.query.wallets.findFirst({
          where: eq(schema.wallets.userId, userId),
        });

        if (!userWallet) {
          throw new NotFoundException(`Wallet not found for user ID: ${userId}`);
        }

        if(!user?.business){
            throw new NotFoundException(`No business found for the user ID: ${userId}`);
        }

        let cost: number;
        let endDate: Date;
        const now = fromZonedTime(new Date(), 'UTC');

        switch (plan) {
          case SubscriptionPlan.Monthly:
            cost = 50.0;
            endDate = addMonths(now, 1);
            break;
          case SubscriptionPlan.Yearly:
            cost = 10 * 50.0;
            endDate = addYears(now, 1);
            break;
          default:
            throw new BadRequestException('Invalid subscription plan.');
        }

        if (parseFloat(userWallet.balance) < cost) {
          throw new BadRequestException('Insufficient wallet balance.');
        }

        const newBalance = parseFloat(userWallet.balance) - cost;

        // Update wallet balance
        await tx.update(schema.wallets)
          .set({
            balance: newBalance.toFixed(2),
            updatedAt: now,
            updatedBy: currentUser.fullName,
          })
          .where(eq(schema.wallets.walletId, userWallet.walletId));

        // Record transaction
        const transactionId = `phi${generateId()}`;
        await tx.insert(schema.transactions).values({
          transactionId,
          senderWalletId: userWallet.walletId,
          recipientWalletId: userWallet.walletId, // Self-transaction for subscription
          amount: cost.toFixed(2),
          currency: userWallet.currency || 'USD',
          type: 'payment',
          status: 'completed',
          description: `Subscription to ${plan} plan`,
          createdBy: currentUser.fullName,
          createdAt: now,
          updatedBy: currentUser.fullName,
          updatedAt: now,
        });

        // Create or update subscription
        const existingSubscription = await tx.query.subscriptions.findFirst({
          where: eq(schema.subscriptions.userId, userId),
        });

        if (existingSubscription) {
          await tx.update(schema.subscriptions)
            .set({
              plan,
              status: SubscriptionStatus.Active,
              startDate: now,
              endDate,
              updatedBy: currentUser.fullName,
              updatedAt: now,
            })
            .where(eq(schema.subscriptions.subscriptionId, existingSubscription.subscriptionId));
        } else {
          const subscriptionId = `phi${generateId()}`;
          await tx.insert(schema.subscriptions).values({
            subscriptionId,
            userId,
            businessId: user.business.businessId, // Set to null if not provided
            plan,
            status: SubscriptionStatus.Active,
            startDate: now,
            endDate,
            createdBy: currentUser.fullName,
            createdAt: now,
            updatedBy: currentUser.fullName,
            updatedAt: now,
          });
        }
        return { message: `Successfully subscribed to ${plan} plan.` };
      });
    } catch (error) {
      this.logger.error(
        `Error subscribing to plan: ${error.message}`,
        error.stack,
      );
      this.databaseErrorService.handleDrizzleError(error, 'subscribe to plan');
    }
  }
}
